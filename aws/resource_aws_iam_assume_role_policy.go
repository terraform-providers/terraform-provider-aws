package aws

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsIamAssumeRolePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamAssumeRolePolicyCreate,
		Read:   resourceAwsIamAssumeRolePolicyRead,
		Update: resourceAwsIamAssumeRolePolicyUpdate,
		Delete: resourceAwsIamAssumeRolePolicyDelete,

		Schema: map[string]*schema.Schema{
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.ValidateJsonString,
			},
		},
	}
}

func resourceAwsIamAssumeRolePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	assumeRolePolicyInput := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(d.Get("role_name").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
	}
	_, err := iamconn.UpdateAssumeRolePolicy(assumeRolePolicyInput)
	if err != nil {
		return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
	}

	d.SetId(d.Get("role_name").(string))

	return resourceAwsIamAssumeRolePolicyRead(d, meta)
}

func resourceAwsIamAssumeRolePolicyRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.GetRoleInput{
		RoleName: aws.String(d.Id()),
	}

	getResp, err := iamconn.GetRole(request)
	if err != nil {
		return fmt.Errorf("Error reading IAM Role %s: %s", d.Id(), err)
	}

	role := getResp.Role

	if err := d.Set("role_name", role.RoleName); err != nil {
		return err
	}

	assumeRolePolicy, err := url.QueryUnescape(*role.AssumeRolePolicyDocument)
	if err != nil {
		return err
	}
	if err := d.Set("policy", assumeRolePolicy); err != nil {
		return err
	}
	return nil
}

func resourceAwsIamAssumeRolePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if d.HasChange("policy") {
		assumeRolePolicyInput := &iam.UpdateAssumeRolePolicyInput{
			RoleName:       aws.String(d.Get("role_name").(string)),
			PolicyDocument: aws.String(d.Get("policy").(string)),
		}
		_, err := iamconn.UpdateAssumeRolePolicy(assumeRolePolicyInput)
		if err != nil {
			return fmt.Errorf("Error Updating IAM Role (%s) Assume Role Policy: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsIamAssumeRolePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	// No need to do anything, assume role policy will be deleted when the role is deleted
	return nil
}
