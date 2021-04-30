package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/waiter"
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
		return fmt.Errorf("error associating principal with RAM resource share: %w", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareArn, principal))

	// AWS Account ID Principals need to be accepted to become ASSOCIATED
	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		return resourceAwsRamPrincipalAssociationRead(d, meta)
	}

	if _, err := waiter.ResourceSharePrincipalAssociated(conn, resourceShareArn, principal); err != nil {
		return fmt.Errorf("error waiting for RAM principal association (%s) to become ready: %w", d.Id(), err)
	}

	return resourceAwsRamPrincipalAssociationRead(d, meta)
}

func resourceAwsRamPrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn, principal, err := resourceAwsRamPrincipalAssociationParseId(d.Id())
	if err != nil {
		return fmt.Errorf("error reading RAM Principal Association, parsing ID (%s): %w", d.Id(), err)
	}

	var association *ram.ResourceShareAssociation

	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		// AWS Account ID Principals need to be accepted to become ASSOCIATED
		association, err = finder.ResourceSharePrincipalAssociationByShareARNPrincipal(conn, resourceShareArn, principal)
	} else {
		association, err = waiter.ResourceSharePrincipalAssociated(conn, resourceShareArn, principal)
	}

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException) || tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException)) {
		log.Printf("[WARN] No RAM resource share principal association with ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if association == nil || aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusDisassociated {
		log.Printf("[WARN] RAM resource share principal association with ARN (%s) found, but empty or disassociated - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociating {
		return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s), status not associating or associated: %s", resourceShareArn, principal, aws.StringValue(association.Status))
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

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if _, err := waiter.ResourceSharePrincipalDisassociated(conn, resourceShareArn, principal); err != nil {
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
