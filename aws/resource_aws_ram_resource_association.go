package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsRamResourceAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceAssociationCreate,
		Read:   resourceAwsRamResourceAssociationRead,
		Delete: resourceAwsRamResourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_share_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsRamResourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn
	resourceARN := d.Get("resource_arn").(string)
	resourceShareARN := d.Get("resource_share_arn").(string)

	input := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Associating RAM Resource Share: %s", input)
	_, err := conn.AssociateResourceShare(input)
	if err != nil {
		return fmt.Errorf("error associating RAM Resource Share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareARN, resourceARN))

	if err := waitForRamResourceShareResourceAssociation(conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	return resourceAwsRamResourceAssociationRead(d, meta)
}

func resourceAwsRamResourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareARN, resourceARN, err := decodeRamResourceAssociationID(d.Id())
	if err != nil {
		return err
	}

	resourceShareAssociation, err := getRamResourceShareAssociation(conn, resourceShareARN, resourceARN)

	if err != nil {
		return fmt.Errorf("error reading RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	if resourceShareAssociation == nil {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not found, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return nil
	}

	if aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not associated, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return nil
	}

	d.Set("resource_arn", resourceARN)
	d.Set("resource_share_arn", resourceShareARN)

	return nil
}

func resourceAwsRamResourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareARN, resourceARN, err := decodeRamResourceAssociationID(d.Id())
	if err != nil {
		return err
	}

	input := &ram.DisassociateResourceShareInput{
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Disassociating RAM Resource Share: %s", input)
	_, err = conn.DisassociateResourceShare(input)

	if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	if err := waitForRamResourceShareResourceDisassociation(conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Resource Association (%s) disassociation: %s", resourceShareARN, resourceARN, err)
	}

	return nil
}

func decodeRamResourceAssociationID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,RESOURCE", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}

func getRamResourceShareAssociation(conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypeResource),
		ResourceArn:       aws.String(resourceARN),
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

func ramResourceAssociationStateRefreshFunc(conn *ram.RAM, resourceShareARN, resourceARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resourceShareAssociation, err := getRamResourceShareAssociation(conn, resourceShareARN, resourceARN)

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if resourceShareAssociation == nil {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		if aws.StringValue(resourceShareAssociation.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(resourceShareAssociation.StatusMessage))
			return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), extendedErr
		}

		return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), nil
	}
}

func waitForRamResourceShareResourceAssociation(conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: ramResourceAssociationStateRefreshFunc(conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForRamResourceShareResourceDisassociation(conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: ramResourceAssociationStateRefreshFunc(conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}
