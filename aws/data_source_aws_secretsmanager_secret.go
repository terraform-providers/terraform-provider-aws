package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSecretsManagerSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSecretsManagerSecretRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rotation_enabled": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation data source instead",
				Type:       schema.TypeBool,
				Computed:   true,
			},
			"rotation_lambda_arn": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation data source instead",
				Type:       schema.TypeString,
				Computed:   true,
			},
			"rotation_rules": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation data source instead",
				Type:       schema.TypeList,
				Computed:   true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsSecretsManagerSecretRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var secretID string
	if v, ok := d.GetOk("arn"); ok {
		secretID = v.(string)
	}
	if v, ok := d.GetOk("name"); ok {
		if secretID != "" {
			return errors.New("specify only arn or name")
		}
		secretID = v.(string)
	}

	if secretID == "" {
		return errors.New("must specify either arn or name")
	}

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretID),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret: %s", input)
	output, err := conn.DescribeSecret(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("Secrets Manager Secret %q not found", secretID)
		}
		return fmt.Errorf("error reading Secrets Manager Secret: %s", err)
	}

	if output.ARN == nil {
		return fmt.Errorf("Secrets Manager Secret %q not found", secretID)
	}

	d.SetId(aws.StringValue(output.ARN))
	d.Set("arn", output.ARN)
	d.Set("description", output.Description)
	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("name", output.Name)
	d.Set("rotation_enabled", output.RotationEnabled)
	d.Set("rotation_lambda_arn", output.RotationLambdaARN)
	d.Set("policy", "")

	pIn := &secretsmanager.GetResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Secrets Manager Secret policy: %s", pIn)
	pOut, err := conn.GetResourcePolicy(pIn)
	if err != nil {
		return fmt.Errorf("error reading Secrets Manager Secret policy: %s", err)
	}

	if pOut != nil && pOut.ResourcePolicy != nil {
		policy, err := structure.NormalizeJsonString(aws.StringValue(pOut.ResourcePolicy))
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		d.Set("policy", policy)
	}

	if err := d.Set("rotation_rules", flattenSecretsManagerRotationRules(output.RotationRules)); err != nil {
		return fmt.Errorf("error setting rotation_rules: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.SecretsmanagerKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
