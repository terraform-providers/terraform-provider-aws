package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsRamPrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamPrincipalAssociationCreate,
		Read:   resourceAwsRamPrincipalAssociationRead,
		Delete: resourceAwsRamPrincipalAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validateAwsAccountId,
					validateArn,
				),
			},
		},
	}
}

func resourceAwsRamPrincipalAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn := d.Get("resource_share_arn").(string)
	principal := d.Get("principal").(string)

	request := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Create RAM principal association request:", request)
	_, err := conn.AssociateResourceShare(request)
	if err != nil {
		return fmt.Errorf("Error associating principal with RAM resource share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareArn, principal))

	// AWS Account ID Principals need to be accepted to become ASSOCIATED
	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		return resourceAwsRamPrincipalAssociationRead(d, meta)
	}

	if err := waitForRamResourceSharePrincipalAssociation(conn, resourceShareArn, principal); err != nil {
		return fmt.Errorf("Error waiting for RAM principal association (%s) to become ready: %s", d.Id(), err)
	}

	return resourceAwsRamPrincipalAssociationRead(d, meta)
}

func resourceAwsRamPrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn, principal, err := resourceAwsRamPrincipalAssociationParseId(d.Id())
	if err != nil {
		return err
	}

	resourceShareAssociation, err := getRamResourceSharePrincipalAssociation(conn, resourceShareArn, principal)

	if err != nil {
		return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if resourceShareAssociation == nil {
		log.Printf("[WARN] RAM Resource Share (%s) Principal Association (%s) not found, removing from state", resourceShareArn, principal)
		d.SetId("")
		return nil
	}

	if aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociating {
		log.Printf("[WARN] RAM Resource Share (%s) Principal Association (%s) not associating or associated, removing from state", resourceShareArn, principal)
		d.SetId("")
		return nil
	}

	d.Set("resource_share_arn", resourceShareArn)
	d.Set("principal", principal)

	return nil
}

func resourceAwsRamPrincipalAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn, principal, err := resourceAwsRamPrincipalAssociationParseId(d.Id())
	if err != nil {
		return err
	}

	request := &ram.DisassociateResourceShareInput{
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Delete RAM principal association request:", request)
	_, err = conn.DisassociateResourceShare(request)

	if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if err := waitForRamResourceSharePrincipalDisassociation(conn, resourceShareArn, principal); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Principal Association (%s) disassociation: %s", resourceShareArn, principal, err)
	}

	return nil
}

func resourceAwsRamPrincipalAssociationParseId(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,PRINCIPAL", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}

func getRamResourceSharePrincipalAssociation(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
		Principal:         aws.String(principal),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	output, err := conn.GetResourceShareAssociations(input)

	if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ResourceShareAssociations) == 0 || output.ResourceShareAssociations[0] == nil {
		return nil, nil
	}

	return output.ResourceShareAssociations[0], nil
}

func resourceAwsRamPrincipalAssociationStateRefreshFunc(conn *ram.RAM, resourceShareArn, principal string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		association, err := getRamResourceSharePrincipalAssociation(conn, resourceShareArn, principal)

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if association == nil {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		if aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(association.StatusMessage))
			return association, aws.StringValue(association.Status), extendedErr
		}

		return association, aws.StringValue(association.Status), nil
	}
}

func waitForRamResourceSharePrincipalAssociation(conn *ram.RAM, resourceShareARN, principal string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: resourceAwsRamPrincipalAssociationStateRefreshFunc(conn, resourceShareARN, principal),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForRamResourceSharePrincipalDisassociation(conn *ram.RAM, resourceShareARN, principal string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAwsRamPrincipalAssociationStateRefreshFunc(conn, resourceShareARN, principal),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}
