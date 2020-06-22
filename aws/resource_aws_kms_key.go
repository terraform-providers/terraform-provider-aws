package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/waiter"
)

func resourceAwsKmsKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKmsKeyCreate,
		Read:   resourceAwsKmsKeyRead,
		Update: resourceAwsKmsKeyUpdate,
		Delete: resourceAwsKmsKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  kms.KeyUsageTypeEncryptDecrypt,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					kms.KeyUsageTypeEncryptDecrypt,
					kms.KeyUsageTypeSignVerify,
				}, false),
			},
			"customer_master_key_spec": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  kms.CustomerMasterKeySpecSymmetricDefault,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					kms.CustomerMasterKeySpecSymmetricDefault,
					kms.CustomerMasterKeySpecRsa2048,
					kms.CustomerMasterKeySpecRsa3072,
					kms.CustomerMasterKeySpecRsa4096,
					kms.CustomerMasterKeySpecEccNistP256,
					kms.CustomerMasterKeySpecEccNistP384,
					kms.CustomerMasterKeySpecEccNistP521,
					kms.CustomerMasterKeySpecEccSecgP256k1,
				}, false),
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"enable_key_rotation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deletion_window_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsKmsKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	// Allow aws to chose default values if we don't pass them
	req := &kms.CreateKeyInput{
		CustomerMasterKeySpec: aws.String(d.Get("customer_master_key_spec").(string)),
		KeyUsage:              aws.String(d.Get("key_usage").(string)),
	}
	if v, exists := d.GetOk("description"); exists {
		req.Description = aws.String(v.(string))
	}
	if v, exists := d.GetOk("policy"); exists {
		req.Policy = aws.String(v.(string))
	}
	if v, exists := d.GetOk("tags"); exists {
		req.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().KmsTags()
	}

	var resp *kms.CreateKeyOutput
	// AWS requires any principal in the policy to exist before the key is created.
	// The KMS service's awareness of principals is limited by "eventual consistency".
	// They acknowledge this here:
	// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateKey(req)
		if isAWSErr(err, kms.ErrCodeMalformedPolicyDocumentException, "") {
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreateKey(req)
	}
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.KeyMetadata.KeyId))
	d.Set("key_id", resp.KeyMetadata.KeyId)

	if enableKeyRotation := d.Get("enable_key_rotation").(bool); enableKeyRotation {
		if err := updateKmsKeyRotationStatus(conn, d); err != nil {
			return err
		}
	}

	if enabled := d.Get("is_enabled").(bool); !enabled {
		if err := updateKmsKeyStatus(conn, d.Id(), enabled); err != nil {
			return err
		}
	}

	return resourceAwsKmsKeyRead(d, meta)
}

func resourceAwsKmsKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &kms.DescribeKeyInput{
		KeyId: aws.String(d.Id()),
	}

	var resp *kms.DescribeKeyOutput
	var err error
	if d.IsNewResource() {
		var out interface{}
		out, err = retryOnAwsCode(kms.ErrCodeNotFoundException, func() (interface{}, error) {
			return conn.DescribeKey(req)
		})
		resp, _ = out.(*kms.DescribeKeyOutput)
	} else {
		resp, err = conn.DescribeKey(req)
	}
	if err != nil {
		return err
	}
	metadata := resp.KeyMetadata

	if aws.StringValue(metadata.KeyState) == kms.KeyStatePendingDeletion {
		log.Printf("[WARN] Removing KMS key %s because it's already gone", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", metadata.Arn)
	d.Set("key_id", metadata.KeyId)
	d.Set("description", metadata.Description)
	d.Set("key_usage", metadata.KeyUsage)
	d.Set("customer_master_key_spec", metadata.CustomerMasterKeySpec)
	d.Set("is_enabled", metadata.Enabled)

	pOut, err := retryOnAwsCode(kms.ErrCodeNotFoundException, func() (interface{}, error) {
		return conn.GetKeyPolicy(&kms.GetKeyPolicyInput{
			KeyId:      aws.String(d.Id()),
			PolicyName: aws.String("default"),
		})
	})
	if err != nil {
		return err
	}

	p := pOut.(*kms.GetKeyPolicyOutput)
	policy, err := structure.NormalizeJsonString(*p.Policy)
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %s", err)
	}
	d.Set("policy", policy)

	out, err := retryOnAwsCode(kms.ErrCodeNotFoundException, func() (interface{}, error) {
		return conn.GetKeyRotationStatus(&kms.GetKeyRotationStatusInput{
			KeyId: aws.String(d.Id()),
		})
	})
	if err != nil {
		return err
	}
	krs, _ := out.(*kms.GetKeyRotationStatusOutput)
	d.Set("enable_key_rotation", krs.KeyRotationEnabled)

	var tags keyvaluetags.KeyValueTags
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		tags, err = keyvaluetags.KmsListTags(conn, d.Id())

		if d.IsNewResource() && isAWSErr(err, kms.ErrCodeNotFoundException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		tags, err = keyvaluetags.KmsListTags(conn, d.Id())
	}

	if err != nil {
		return fmt.Errorf("error listing tags for KMS Key (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsKmsKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	if d.HasChange("is_enabled") && d.Get("is_enabled").(bool) {
		// Enable before any attributes will be modified
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("enable_key_rotation") {
		if err := updateKmsKeyRotationStatus(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		if err := resourceAwsKmsKeyDescriptionUpdate(conn, d); err != nil {
			return err
		}
	}
	if d.HasChange("policy") {
		if err := resourceAwsKmsKeyPolicyUpdate(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("is_enabled") && !d.Get("is_enabled").(bool) {
		// Only disable when all attributes are modified
		// because we cannot modify disabled keys
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("is_enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.KmsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating KMS Key (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsKmsKeyRead(d, meta)
}

func resourceAwsKmsKeyDescriptionUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	description := d.Get("description").(string)
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] KMS key: %s, update description: %s", keyId, description)

	req := &kms.UpdateKeyDescriptionInput{
		Description: aws.String(description),
		KeyId:       aws.String(keyId),
	}
	_, err := conn.UpdateKeyDescription(req)

	return err
}

func resourceAwsKmsKeyPolicyUpdate(conn *kms.KMS, d *schema.ResourceData) error {
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %s", err)
	}
	keyId := d.Get("key_id").(string)

	log.Printf("[DEBUG] KMS key: %s, update policy: %s", keyId, policy)

	req := &kms.PutKeyPolicyInput{
		KeyId:      aws.String(keyId),
		Policy:     aws.String(policy),
		PolicyName: aws.String("default"),
	}
	_, err = conn.PutKeyPolicy(req)

	return err
}

func updateKmsKeyStatus(conn *kms.KMS, id string, shouldBeEnabled bool) error {
	var err error

	if shouldBeEnabled {
		log.Printf("[DEBUG] Enabling KMS key %q", id)
		_, err = conn.EnableKey(&kms.EnableKeyInput{
			KeyId: aws.String(id),
		})
	} else {
		log.Printf("[DEBUG] Disabling KMS key %q", id)
		_, err = conn.DisableKey(&kms.DisableKeyInput{
			KeyId: aws.String(id),
		})
	}

	if err != nil {
		return fmt.Errorf("Failed to set KMS key %q status to %t: %q",
			id, shouldBeEnabled, err.Error())
	}

	// Wait for propagation since KMS is eventually consistent
	wait := resource.StateChangeConf{
		Pending:                   []string{fmt.Sprintf("%t", !shouldBeEnabled)},
		Target:                    []string{fmt.Sprintf("%t", shouldBeEnabled)},
		Timeout:                   20 * time.Minute,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 15,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if KMS key %s enabled status is %t",
				id, shouldBeEnabled)
			resp, err := conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: aws.String(id),
			})
			if err != nil {
				awsErr, ok := err.(awserr.Error)
				if ok && awsErr.Code() == "NotFoundException" {
					return nil, fmt.Sprintf("%t", !shouldBeEnabled), nil
				}
				return resp, "FAILED", err
			}
			status := fmt.Sprintf("%t", *resp.KeyMetadata.Enabled)
			log.Printf("[DEBUG] KMS key %s status received: %s, retrying", id, status)

			return resp, status, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed setting KMS key status to %t: %s", shouldBeEnabled, err)
	}

	return nil
}

func updateKmsKeyRotationStatus(conn *kms.KMS, d *schema.ResourceData) error {
	shouldEnableRotation := d.Get("enable_key_rotation").(bool)

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		err := handleKeyRotation(conn, shouldEnableRotation, aws.String(d.Id()))

		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "DisabledException" {
				return resource.RetryableError(err)
			}
			if ok && awsErr.Code() == "NotFoundException" {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		err = handleKeyRotation(conn, shouldEnableRotation, aws.String(d.Id()))
	}

	if err != nil {
		return fmt.Errorf("Failed to set key rotation for %q to %t: %q",
			d.Id(), shouldEnableRotation, err.Error())
	}

	// Wait for propagation since KMS is eventually consistent
	wait := resource.StateChangeConf{
		Pending:                   []string{fmt.Sprintf("%t", !shouldEnableRotation)},
		Target:                    []string{fmt.Sprintf("%t", shouldEnableRotation)},
		Timeout:                   5 * time.Minute,
		MinTimeout:                1 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if KMS key %s rotation status is %t",
				d.Id(), shouldEnableRotation)

			out, err := retryOnAwsCode("NotFoundException", func() (interface{}, error) {
				return conn.GetKeyRotationStatus(&kms.GetKeyRotationStatusInput{
					KeyId: aws.String(d.Id()),
				})
			})
			if err != nil {
				return 42, "", err
			}
			resp, _ := out.(*kms.GetKeyRotationStatusOutput)

			status := fmt.Sprintf("%t", *resp.KeyRotationEnabled)
			log.Printf("[DEBUG] KMS key %s rotation status received: %s, retrying", d.Id(), status)

			return resp, status, nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed setting KMS key rotation status to %t: %s", shouldEnableRotation, err)
	}

	return nil
}

func handleKeyRotation(conn *kms.KMS, shouldEnableRotation bool, keyId *string) error {
	var err error
	if shouldEnableRotation {
		log.Printf("[DEBUG] Enabling key rotation for KMS key %q", *keyId)
		_, err = conn.EnableKeyRotation(&kms.EnableKeyRotationInput{
			KeyId: keyId,
		})
	} else {
		log.Printf("[DEBUG] Disabling key rotation for KMS key %q", *keyId)
		_, err = conn.DisableKeyRotation(&kms.DisableKeyRotationInput{
			KeyId: keyId,
		})
	}
	return err
}

func resourceAwsKmsKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	keyId := d.Get("key_id").(string)

	req := &kms.ScheduleKeyDeletionInput{
		KeyId: aws.String(keyId),
	}
	if v, exists := d.GetOk("deletion_window_in_days"); exists {
		req.PendingWindowInDays = aws.Int64(int64(v.(int)))
	}
	_, err := conn.ScheduleKeyDeletion(req)

	if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error scheduling deletion for KMS Key (%s): %w", d.Id(), err)
	}

	_, err = waiter.KeyStatePendingDeletion(conn, d.Id())

	if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for KMS Key (%s) to schedule deletion: %w", d.Id(), err)
	}

	return nil
}
