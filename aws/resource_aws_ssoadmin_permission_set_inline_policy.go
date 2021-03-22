package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSsoAdminPermissionSetInlinePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoAdminPermissionSetInlinePolicyPut,
		Read:   resourceAwsSsoAdminPermissionSetInlinePolicyRead,
		Update: resourceAwsSsoAdminPermissionSetInlinePolicyPut,
		Delete: resourceAwsSsoAdminPermissionSetInlinePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"inline_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsSsoAdminPermissionSetInlinePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	input := &ssoadmin.PutInlinePolicyToPermissionSetInput{
		InlinePolicy:     aws.String(d.Get("inline_policy").(string)),
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := conn.PutInlinePolicyToPermissionSet(input)
	if err != nil {
		return fmt.Errorf("error putting Inline Policy for SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", permissionSetArn, instanceArn))

	// (Re)provision ALL accounts after making the above changes
	if err := provisionSsoAdminPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourceAwsSsoAdminPermissionSetInlinePolicyRead(d, meta)
}

func resourceAwsSsoAdminPermissionSetInlinePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	permissionSetArn, instanceArn, err := parseSsoAdminResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID: %w", err)
	}

	input := &ssoadmin.GetInlinePolicyForPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	output, err := conn.GetInlinePolicyForPermissionSet(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Inline Policy for SSO Permission Set (%s) not found, removing from state", permissionSetArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Inline Policy for SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	if output == nil {
		return fmt.Errorf("error reading Inline Policy for SSO Permission Set (%s): empty output", permissionSetArn)
	}

	d.Set("inline_policy", output.InlinePolicy)
	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)

	return nil
}

func resourceAwsSsoAdminPermissionSetInlinePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	permissionSetArn, instanceArn, err := parseSsoAdminResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Permission Set Inline Policy ID: %w", err)
	}

	input := &ssoadmin.DeleteInlinePolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err = conn.DeleteInlinePolicyFromPermissionSet(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error detaching Inline Policy from SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	return nil
}
