package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/secretsmanager/waiter"
)

func resourceAwsSecretsManagerSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecretsManagerSecretCreate,
		Read:   resourceAwsSecretsManagerSecretRead,
		Update: resourceAwsSecretsManagerSecretUpdate,
		Delete: resourceAwsSecretsManagerSecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateSecretManagerSecretName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateSecretManagerSecretNamePrefix,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"recovery_window_in_days": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(int)
					if value == 0 {
						return
					}
					if value >= 7 && value <= 30 {
						return
					}
					errors = append(errors, fmt.Errorf("%q must be 0 or between 7 and 30", k))
					return
				},
			},
			"rotation_enabled": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation resource instead",
				Type:       schema.TypeBool,
				Computed:   true,
			},
			"rotation_lambda_arn": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation resource instead",
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
			},
			"rotation_rules": {
				Deprecated: "Use the aws_secretsmanager_secret_rotation resource instead",
				Type:       schema.TypeList,
				Computed:   true,
				Optional:   true,
				MaxItems:   1,
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

func resourceAwsSecretsManagerSecretCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	var secretName string
	if v, ok := d.GetOk("name"); ok {
		secretName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		secretName = resource.PrefixedUniqueId(v.(string))
	} else {
		secretName = resource.UniqueId()
	}

	input := &secretsmanager.CreateSecretInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(secretName),
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SecretsmanagerTags()
	}

	if v, ok := d.GetOk("kms_key_id"); ok && v.(string) != "" {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Secrets Manager Secret: %s", input)

	// Retry for secret recreation after deletion
	var output *secretsmanager.CreateSecretOutput
	err := resource.Retry(waiter.DeletionPropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateSecret(input)
		// Temporarily retry on these errors to support immediate secret recreation:
		// InvalidRequestException: You can’t perform this operation on the secret because it was deleted.
		// InvalidRequestException: You can't create this secret because a secret with this name is already scheduled for deletion.
		if isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException, "scheduled for deletion") || isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException, "was deleted") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateSecret(input)
	}
	if err != nil {
		return fmt.Errorf("error creating Secrets Manager Secret: %s", err)
	}

	d.SetId(aws.StringValue(output.ARN))

	if v, ok := d.GetOk("policy"); ok && v.(string) != "" {
		input := &secretsmanager.PutResourcePolicyInput{
			ResourcePolicy: aws.String(v.(string)),
			SecretId:       aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %s", input)
		_, err := conn.PutResourcePolicy(input)
		if err != nil {
			return fmt.Errorf("error setting Secrets Manager Secret %q policy: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
		input := &secretsmanager.RotateSecretInput{
			RotationLambdaARN: aws.String(v.(string)),
			RotationRules:     expandSecretsManagerRotationRules(d.Get("rotation_rules").([]interface{})),
			SecretId:          aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Enabling Secrets Manager Secret rotation: %s", input)
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
			return fmt.Errorf("error enabling Secrets Manager Secret %q rotation: %s", d.Id(), err)
		}
	}

	return resourceAwsSecretsManagerSecretRead(d, meta)
}

func resourceAwsSecretsManagerSecretRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret: %s", input)
	output, err := conn.DescribeSecret(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Secrets Manager Secret %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Secrets Manager Secret: %s", err)
	}

	d.Set("arn", output.ARN)
	d.Set("description", output.Description)
	d.Set("kms_key_id", output.KmsKeyId)
	d.Set("name", output.Name)

	pIn := &secretsmanager.GetResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Secrets Manager Secret policy: %s", pIn)
	pOut, err := conn.GetResourcePolicy(pIn)
	if err != nil {
		return fmt.Errorf("error reading Secrets Manager Secret policy: %s", err)
	}

	if pOut.ResourcePolicy != nil {
		policy, err := structure.NormalizeJsonString(aws.StringValue(pOut.ResourcePolicy))
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		d.Set("policy", policy)
	}

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

	if err := d.Set("tags", keyvaluetags.SecretsmanagerKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSecretsManagerSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	if d.HasChanges("description", "kms_key_id") {
		input := &secretsmanager.UpdateSecretInput{
			Description: aws.String(d.Get("description").(string)),
			SecretId:    aws.String(d.Id()),
		}

		if v, ok := d.GetOk("kms_key_id"); ok && v.(string) != "" {
			input.KmsKeyId = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Secrets Manager Secret: %s", input)
		_, err := conn.UpdateSecret(input)
		if err != nil {
			return fmt.Errorf("error updating Secrets Manager Secret: %s", err)
		}
	}

	if d.HasChange("policy") {
		if v, ok := d.GetOk("policy"); ok && v.(string) != "" {
			policy, err := structure.NormalizeJsonString(v.(string))
			if err != nil {
				return fmt.Errorf("policy contains an invalid JSON: %s", err)
			}
			input := &secretsmanager.PutResourcePolicyInput{
				ResourcePolicy: aws.String(policy),
				SecretId:       aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %s", input)
			_, err = conn.PutResourcePolicy(input)
			if err != nil {
				return fmt.Errorf("error setting Secrets Manager Secret %q policy: %s", d.Id(), err)
			}
		} else {
			input := &secretsmanager.DeleteResourcePolicyInput{
				SecretId: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Removing Secrets Manager Secret policy: %s", input)
			_, err := conn.DeleteResourcePolicy(input)
			if err != nil {
				return fmt.Errorf("error removing Secrets Manager Secret %q policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("rotation_lambda_arn", "rotation_rules") {
		if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
			input := &secretsmanager.RotateSecretInput{
				RotationLambdaARN: aws.String(v.(string)),
				RotationRules:     expandSecretsManagerRotationRules(d.Get("rotation_rules").([]interface{})),
				SecretId:          aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Enabling Secrets Manager Secret rotation: %s", input)
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
				return fmt.Errorf("error updating Secrets Manager Secret %q rotation: %s", d.Id(), err)
			}
		} else {
			input := &secretsmanager.CancelRotateSecretInput{
				SecretId: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Cancelling Secrets Manager Secret rotation: %s", input)
			_, err := conn.CancelRotateSecret(input)
			if err != nil {
				return fmt.Errorf("error cancelling Secret Manager Secret %q rotation: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SecretsmanagerUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSecretsManagerSecretRead(d, meta)
}

func resourceAwsSecretsManagerSecretDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(d.Id()),
	}

	recoveryWindowInDays := d.Get("recovery_window_in_days").(int)
	if recoveryWindowInDays == 0 {
		input.ForceDeleteWithoutRecovery = aws.Bool(true)
	} else {
		input.RecoveryWindowInDays = aws.Int64(int64(recoveryWindowInDays))
	}

	log.Printf("[DEBUG] Deleting Secrets Manager Secret: %s", input)
	_, err := conn.DeleteSecret(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Secrets Manager Secret: %s", err)
	}

	return nil
}
