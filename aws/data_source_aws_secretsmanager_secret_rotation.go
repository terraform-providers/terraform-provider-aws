package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsSecretsManagerSecretRotation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSecretsManagerSecretRotationRead,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 2048),
				Required:     true,
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rotation_rules": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsSecretsManagerSecretRotationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretID),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret: %s", input)
	output, err := conn.DescribeSecret(input)
	if err != nil {
		return fmt.Errorf("error reading Secrets Manager Secret: %s", err)
	}

	if output.ARN == nil {
		return fmt.Errorf("Secrets Manager Secret %q not found", secretID)
	}

	d.SetId(aws.StringValue(output.ARN))
	d.Set("rotation_enabled", output.RotationEnabled)
	d.Set("rotation_lambda_arn", output.RotationLambdaARN)

	if err := d.Set("rotation_rules", flattenSecretsManagerRotationRules(output.RotationRules)); err != nil {
		return fmt.Errorf("error setting rotation_rules: %s", err)
	}

	return nil
}
