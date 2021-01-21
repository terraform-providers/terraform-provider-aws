package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsEbsDefaultKmsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEbsDefaultKmsKeyCreate,
		Read:   resourceAwsEbsDefaultKmsKeyRead,
		Delete: resourceAwsEbsDefaultKmsKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsEbsDefaultKmsKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.ModifyEbsDefaultKmsKeyId(&ec2.ModifyEbsDefaultKmsKeyIdInput{
		KmsKeyId: aws.String(d.Get("key_arn").(string)),
	})
	if err != nil {
		return fmt.Errorf("error creating EBS default KMS key: %s", err)
	}

	d.SetId(aws.StringValue(resp.KmsKeyId))

	return resourceAwsEbsDefaultKmsKeyRead(d, meta)
}

func resourceAwsEbsDefaultKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error reading EBS default KMS key: %s", err)
	}

	d.Set("key_arn", resp.KmsKeyId)

	return nil
}

func resourceAwsEbsDefaultKmsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.ResetEbsDefaultKmsKeyId(&ec2.ResetEbsDefaultKmsKeyIdInput{})
	if err != nil {
		return fmt.Errorf("error deleting EBS default KMS key: %s", err)
	}

	return nil
}
