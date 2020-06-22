package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSecretsManagerSecretRotation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecretsManagerSecretRotationCreate,
		Read:   resourceAwsSecretsManagerSecretRotationRead,
		Update: resourceAwsSecretsManagerSecretRotationUpdate,
		Delete: resourceAwsSecretsManagerSecretRotationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rotation_rules": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSecretsManagerSecretRotationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	secretID := d.Get("secret_id").(string)

	if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
		input := &secretsmanager.RotateSecretInput{
			RotationLambdaARN: aws.String(v.(string)),
			RotationRules:     expandSecretsManagerRotationRules(d.Get("rotation_rules").([]interface{})),
			SecretId:          aws.String(secretID),
		}

		log.Printf("[DEBUG] Enabling Secrets Manager Secret rotation: %s", input)
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			output, err := conn.RotateSecret(input)
			if err != nil {
				// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
				if isAWSErr(err, "AccessDeniedException", "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			d.SetId(aws.StringValue(output.ARN))

			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.RotateSecret(input)
		}

		if err != nil {
			return fmt.Errorf("error enabling Secrets Manager Secret %q rotation: %s", d.Id(), err)
		}
	}

	return resourceAwsSecretsManagerSecretRotationRead(d, meta)
}

func resourceAwsSecretsManagerSecretRotationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret Rotation: %s", input)
	output, err := conn.DescribeSecret(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Secrets Manager Secret Rotation %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Secrets Manager Secret Rotation: %s", err)
	}

	d.Set("secret_id", d.Id())
	d.Set("rotation_enabled", output.RotationEnabled)

	if aws.BoolValue(output.RotationEnabled) {
		d.Set("rotation_lambda_arn", output.RotationLambdaARN)
		if err := d.Set("rotation_rules", flattenSecretsManagerRotationRules(output.RotationRules)); err != nil {
			return fmt.Errorf("error setting rotation_rules: %s", err)
		}
	} else {
		d.Set("rotation_lambda_arn", "")
		d.Set("rotation_rules", []interface{}{})
	}

	return nil
}

func resourceAwsSecretsManagerSecretRotationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	secretID := d.Get("secret_id").(string)

	if d.HasChanges("rotation_lambda_arn", "rotation_rules") {
		if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
			input := &secretsmanager.RotateSecretInput{
				RotationLambdaARN: aws.String(v.(string)),
				RotationRules:     expandSecretsManagerRotationRules(d.Get("rotation_rules").([]interface{})),
				SecretId:          aws.String(secretID),
			}

			log.Printf("[DEBUG] Enabling Secrets Manager Secret Rotation: %s", input)
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.RotateSecret(input)
				if err != nil {
					// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
					if isAWSErr(err, "AccessDeniedException", "") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if isResourceTimeoutError(err) {
				_, err = conn.RotateSecret(input)
			}

			if err != nil {
				return fmt.Errorf("error updating Secrets Manager Secret Rotation %q : %s", d.Id(), err)
			}
		} else {
			input := &secretsmanager.CancelRotateSecretInput{
				SecretId: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Cancelling Secrets Manager Secret Rotation: %s", input)
			_, err := conn.CancelRotateSecret(input)
			if err != nil {
				return fmt.Errorf("error cancelling Secret Manager Secret Rotation %q : %s", d.Id(), err)
			}
		}
	}

	return resourceAwsSecretsManagerSecretRotationRead(d, meta)
}

func resourceAwsSecretsManagerSecretRotationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.CancelRotateSecretInput{
		SecretId: aws.String(secretID),
	}

	log.Printf("[DEBUG] Deleting Secrets Manager Rotation: %s", input)
	_, err := conn.CancelRotateSecret(input)
	if err != nil {
		return fmt.Errorf("error cancelling Secret Manager Secret %q rotation: %s", d.Id(), err)
	}

	return nil
}

func expandSecretsManagerRotationRules(l []interface{}) *secretsmanager.RotationRulesType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	rules := &secretsmanager.RotationRulesType{
		AutomaticallyAfterDays: aws.Int64(int64(m["automatically_after_days"].(int))),
	}

	return rules
}

func flattenSecretsManagerRotationRules(rules *secretsmanager.RotationRulesType) []interface{} {
	if rules == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"automatically_after_days": int(aws.Int64Value(rules.AutomaticallyAfterDays)),
	}

	return []interface{}{m}
}
