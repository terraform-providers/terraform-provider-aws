package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsShieldProtection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsShieldProtectionCreate,
		Read:   resourceAwsShieldProtectionRead,
		Delete: resourceAwsShieldProtectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsShieldProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.CreateProtectionInput{
		Name:        aws.String(d.Get("name").(string)),
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	resp, err := conn.CreateProtection(input)
	if err != nil {
		return fmt.Errorf("error creating Shield Protection: %s", err)
	}
	d.SetId(*resp.ProtectionId)
	return resourceAwsShieldProtectionRead(d, meta)
}

func resourceAwsShieldProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtection(input)
	if err != nil {
		return fmt.Errorf("error reading Shield Protection (%s): %s", d.Id(), err)
	}
	d.Set("name", resp.Protection.Name)
	d.Set("resource_arn", resp.Protection.ResourceArn)
	return nil
}

func resourceAwsShieldProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteProtection(input)

	if isAWSErr(err, shield.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Shield Protection (%s): %s", d.Id(), err)
	}
	return nil
}
