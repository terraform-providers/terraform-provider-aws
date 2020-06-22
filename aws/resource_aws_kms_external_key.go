package aws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/waiter"
)

func resourceAwsKmsExternalKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKmsExternalKeyCreate,
		Read:   resourceAwsKmsExternalKeyRead,
		Update: resourceAwsKmsExternalKeyUpdate,
		Delete: resourceAwsKmsExternalKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_window_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(7, 30),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 8192),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"expiration_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_material_base64": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"key_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 32768),
					validation.StringIsJSON,
				),
			},
			"tags": tagsSchema(),
			"valid_to": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
		},
	}
}

func resourceAwsKmsExternalKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	input := &kms.CreateKeyInput{
		KeyUsage: aws.String(kms.KeyUsageTypeEncryptDecrypt),
		Origin:   aws.String(kms.OriginTypeExternal),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok {
		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().KmsTags()
	}

	var output *kms.CreateKeyOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateKey(input)

		// AWS requires any principal in the policy to exist before the key is created.
		// The KMS service's awareness of principals is limited by "eventual consistency".
		// KMS will report this error until it can validate the policy itself.
		// They acknowledge this here:
		// http://docs.aws.amazon.com/kms/latest/APIReference/API_CreateKey.html
		if isAWSErr(err, kms.ErrCodeMalformedPolicyDocumentException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateKey(input)
	}

	if err != nil {
		return fmt.Errorf("error creating KMS External Key: %s", err)
	}

	d.SetId(aws.StringValue(output.KeyMetadata.KeyId))

	if v, ok := d.GetOk("key_material_base64"); ok {
		if err := importKmsExternalKeyMaterial(conn, d.Id(), v.(string), d.Get("valid_to").(string)); err != nil {
			return fmt.Errorf("error importing KMS External Key (%s) material: %s", d.Id(), err)
		}

		if !d.Get("enabled").(bool) {
			if err := updateKmsKeyStatus(conn, d.Id(), false); err != nil {
				return fmt.Errorf("error disabling KMS External Key (%s): %s", d.Id(), err)
			}
		}
	}

	return resourceAwsKmsExternalKeyRead(d, meta)
}

func resourceAwsKmsExternalKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &kms.DescribeKeyInput{
		KeyId: aws.String(d.Id()),
	}

	var output *kms.DescribeKeyOutput
	// Retry for KMS eventual consistency on creation
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error

		output, err = conn.DescribeKey(input)

		if d.IsNewResource() && isAWSErr(err, kms.ErrCodeNotFoundException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.DescribeKey(input)
	}

	if !d.IsNewResource() && isAWSErr(err, kms.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] KMS External Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing KMS External Key (%s): %s", d.Id(), err)
	}

	if output == nil || output.KeyMetadata == nil {
		return fmt.Errorf("error describing KMS External Key (%s): empty response", d.Id())
	}

	metadata := output.KeyMetadata

	if aws.StringValue(metadata.KeyState) == kms.KeyStatePendingDeletion {
		log.Printf("[WARN] KMS External Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	getKeyPolicyInput := &kms.GetKeyPolicyInput{
		KeyId:      metadata.KeyId,
		PolicyName: aws.String("default"),
	}

	getKeyPolicyOutput, err := conn.GetKeyPolicy(getKeyPolicyInput)

	if err != nil {
		return fmt.Errorf("error getting KMS External Key (%s) policy: %s", d.Id(), err)
	}

	if getKeyPolicyOutput == nil {
		return fmt.Errorf("error getting KMS External Key (%s) policy: empty response", d.Id())
	}

	policy, err := structure.NormalizeJsonString(aws.StringValue(getKeyPolicyOutput.Policy))

	if err != nil {
		return fmt.Errorf("error normalizing KMS External Key (%s) policy: %s", d.Id(), err)
	}

	d.Set("arn", metadata.Arn)
	d.Set("description", metadata.Description)
	d.Set("enabled", metadata.Enabled)
	d.Set("expiration_model", metadata.ExpirationModel)
	d.Set("key_state", metadata.KeyState)
	d.Set("key_usage", metadata.KeyUsage)
	d.Set("policy", policy)

	tags, err := keyvaluetags.KmsListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for KMS Key (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("valid_to", "")
	if metadata.ValidTo != nil {
		d.Set("valid_to", aws.TimeValue(metadata.ValidTo).Format(time.RFC3339))
	}

	return nil
}

func resourceAwsKmsExternalKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	if d.HasChange("enabled") && d.Get("enabled").(bool) && d.Get("key_state") != kms.KeyStatePendingImport {
		// Enable before any attributes will be modified
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("description") {
		input := &kms.UpdateKeyDescriptionInput{
			Description: aws.String(d.Get("description").(string)),
			KeyId:       aws.String(d.Id()),
		}

		if _, err := conn.UpdateKeyDescription(input); err != nil {
			return fmt.Errorf("error updating KMS External Key (%s) description: %s", d.Id(), err)
		}
	}

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

		if err != nil {
			return fmt.Errorf("error parsing KMS External Key (%s) policy JSON: %s", d.Id(), err)
		}

		input := &kms.PutKeyPolicyInput{
			KeyId:      aws.String(d.Id()),
			Policy:     aws.String(policy),
			PolicyName: aws.String("default"),
		}

		if _, err := conn.PutKeyPolicy(input); err != nil {
			return fmt.Errorf("error updating KMS External Key (%s) policy: %s", d.Id(), err)
		}
	}

	if d.HasChange("valid_to") {
		if err := importKmsExternalKeyMaterial(conn, d.Id(), d.Get("key_material_base64").(string), d.Get("valid_to").(string)); err != nil {
			return fmt.Errorf("error importing KMS External Key (%s) material: %s", d.Id(), err)
		}
	}

	if d.HasChange("enabled") && !d.Get("enabled").(bool) && d.Get("key_state") != kms.KeyStatePendingImport {
		// Only disable when all attributes are modified
		// because we cannot modify disabled keys
		if err := updateKmsKeyStatus(conn, d.Id(), d.Get("enabled").(bool)); err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.KmsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating KMS External Key (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsKmsExternalKeyRead(d, meta)
}

func resourceAwsKmsExternalKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kmsconn

	input := &kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(d.Id()),
		PendingWindowInDays: aws.Int64(int64(d.Get("deletion_window_in_days").(int))),
	}

	_, err := conn.ScheduleKeyDeletion(input)

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

func importKmsExternalKeyMaterial(conn *kms.KMS, keyID, keyMaterialBase64, validTo string) error {
	getParametersForImportInput := &kms.GetParametersForImportInput{
		KeyId:             aws.String(keyID),
		WrappingAlgorithm: aws.String(kms.AlgorithmSpecRsaesOaepSha256),
		WrappingKeySpec:   aws.String(kms.WrappingKeySpecRsa2048),
	}

	var getParametersForImportOutput *kms.GetParametersForImportOutput
	// Handle KMS eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		getParametersForImportOutput, err = conn.GetParametersForImport(getParametersForImportInput)

		if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		getParametersForImportOutput, err = conn.GetParametersForImport(getParametersForImportInput)
	}

	if err != nil {
		return fmt.Errorf("error getting parameters for import: %s", err)
	}

	if getParametersForImportOutput == nil {
		return fmt.Errorf("error getting parameters for import: empty response")
	}

	keyMaterial, err := base64.StdEncoding.DecodeString(keyMaterialBase64)

	if err != nil {
		return fmt.Errorf("error Base64 decoding key material: %s", err)
	}

	publicKey, err := x509.ParsePKIXPublicKey(getParametersForImportOutput.PublicKey)

	if err != nil {
		return fmt.Errorf("error parsing public key: %s", err)
	}

	encryptedKeyMaterial, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey.(*rsa.PublicKey), keyMaterial, []byte{})

	if err != nil {
		return fmt.Errorf("error encrypting key material: %s", err)
	}

	importKeyMaterialInput := &kms.ImportKeyMaterialInput{
		EncryptedKeyMaterial: encryptedKeyMaterial,
		ExpirationModel:      aws.String(kms.ExpirationModelTypeKeyMaterialDoesNotExpire),
		ImportToken:          getParametersForImportOutput.ImportToken,
		KeyId:                aws.String(keyID),
	}

	if validTo != "" {
		t, err := time.Parse(time.RFC3339, validTo)

		if err != nil {
			return fmt.Errorf("error parsing valid to timestamp: %s", err)
		}

		importKeyMaterialInput.ExpirationModel = aws.String(kms.ExpirationModelTypeKeyMaterialExpires)
		importKeyMaterialInput.ValidTo = aws.Time(t)
	}

	// Handle KMS eventual consistency
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.ImportKeyMaterial(importKeyMaterialInput)

		if isAWSErr(err, kms.ErrCodeNotFoundException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.ImportKeyMaterial(importKeyMaterialInput)
	}

	if err != nil {
		return fmt.Errorf("error importing key material: %s", err)
	}

	return nil
}
