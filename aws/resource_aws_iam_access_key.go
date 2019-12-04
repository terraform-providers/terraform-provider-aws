package aws

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform-plugin-sdk/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsIamAccessKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamAccessKeyCreate,
		Read:   resourceAwsIamAccessKeyRead,
		Update: resourceAwsIamAccessKeyUpdate,
		Delete: resourceAwsIamAccessKeyDelete,

		DeprecationMessage: "AWS SigV2 for SES SMTP passwords is deprecated and 'ses_smtp_password' will be removed.\nUse 'ses_smtp_regions' and 'ses_smtp_passwords' for region-specific AWS SigV4 signed SES SMTP passwords instead.",

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					iam.StatusTypeActive,
					iam.StatusTypeInactive,
				}, false),
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"ses_smtp_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"ses_smtp_passwords": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ses_smtp_regions": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// https://docs.aws.amazon.com/general/latest/gr/rande.html#ses_region
					ValidateFunc: validation.StringInSlice([]string{
						"ap-south-1",
						"ap-southeast-2",
						"eu-central-1",
						"eu-west-1",
						"us-east-1",
						"us-west-2",
					}, true),
				},
			},
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIamAccessKeyCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.CreateAccessKeyInput{
		UserName: aws.String(d.Get("user").(string)),
	}

	createResp, err := iamconn.CreateAccessKey(request)
	if err != nil {
		return fmt.Errorf(
			"Error creating access key for user %s: %s",
			*request.UserName,
			err,
		)
	}

	d.SetId(*createResp.AccessKey.AccessKeyId)

	if createResp.AccessKey == nil || createResp.AccessKey.SecretAccessKey == nil {
		return fmt.Errorf("CreateAccessKey response did not contain a Secret Access Key as expected")
	}

	if v, ok := d.GetOk("pgp_key"); ok {
		pgpKey := v.(string)
		encryptionKey, err := encryption.RetrieveGPGKey(pgpKey)
		if err != nil {
			return err
		}
		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, *createResp.AccessKey.SecretAccessKey, "IAM Access Key Secret")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret", encrypted)
	} else {
		if err := d.Set("secret", createResp.AccessKey.SecretAccessKey); err != nil {
			return err
		}
	}

	// AWS SigV2
	sesSMTPPassword, err := sesSmtpPasswordFromSecretKeyV2(createResp.AccessKey.SecretAccessKey)
	if err != nil {
		return fmt.Errorf("error getting SES SMTP Password from Secret Access Key: %s", err)
	}
	d.Set("ses_smtp_password", sesSMTPPassword)

	// AWS SigV4
	if v, ok := d.GetOk("ses_smtp_regions"); ok && len(v.([]interface{})) > 0 {
		var sesSmtpPasswords []map[string]string

		for _, region := range v.([]interface{}) {
			password, err := sesSmtpPasswordFromSecretKeyV4(createResp.AccessKey.SecretAccessKey, region.(string))
			if err != nil {
				return fmt.Errorf("error getting SES SMTP Password from Secret Access Key: %s", err)
			}
			sesSmtpPasswords = append(
				sesSmtpPasswords,
				map[string]string{
					"region": region.(string),
					"secret": password,
				})
		}
		d.Set("ses_smtp_passwords", sesSmtpPasswords)
	} else {
		d.Set("ses_smtp_passwords", nil)
	}

	return resourceAwsIamAccessKeyReadResult(d, &iam.AccessKeyMetadata{
		AccessKeyId: createResp.AccessKey.AccessKeyId,
		CreateDate:  createResp.AccessKey.CreateDate,
		Status:      createResp.AccessKey.Status,
		UserName:    createResp.AccessKey.UserName,
	})
}

func resourceAwsIamAccessKeyRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.ListAccessKeysInput{
		UserName: aws.String(d.Get("user").(string)),
	}

	getResp, err := iamconn.ListAccessKeys(request)
	if err != nil {
		if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" { // XXX TEST ME
			// the user does not exist, so the key can't exist.
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM access key: %s", err)
	}

	for _, key := range getResp.AccessKeyMetadata {
		if key.AccessKeyId != nil && *key.AccessKeyId == d.Id() {
			return resourceAwsIamAccessKeyReadResult(d, key)
		}
	}

	// Guess the key isn't around anymore.
	d.SetId("")
	return nil
}

func resourceAwsIamAccessKeyReadResult(d *schema.ResourceData, key *iam.AccessKeyMetadata) error {
	d.SetId(*key.AccessKeyId)
	if err := d.Set("user", key.UserName); err != nil {
		return err
	}
	if err := d.Set("status", key.Status); err != nil {
		return err
	}
	return nil
}

func resourceAwsIamAccessKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	if d.HasChange("status") {
		if err := resourceAwsIamAccessKeyStatusUpdate(iamconn, d); err != nil {
			return err
		}
	}

	return resourceAwsIamAccessKeyRead(d, meta)
}

func resourceAwsIamAccessKeyDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		UserName:    aws.String(d.Get("user").(string)),
	}

	if _, err := iamconn.DeleteAccessKey(request); err != nil {
		return fmt.Errorf("Error deleting access key %s: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsIamAccessKeyStatusUpdate(iamconn *iam.IAM, d *schema.ResourceData) error {
	request := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(d.Id()),
		Status:      aws.String(d.Get("status").(string)),
		UserName:    aws.String(d.Get("user").(string)),
	}

	if _, err := iamconn.UpdateAccessKey(request); err != nil {
		return fmt.Errorf("Error updating access key %s: %s", d.Id(), err)
	}
	return nil
}

func hmacSignature(key []byte, value []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(value); err != nil {
		return []byte(""), err
	}
	return h.Sum(nil), nil
}

func sesSmtpPasswordFromSecretKeyV4(key *string, region string) (string, error) {
	if key == nil {
		return "", nil
	}
	version := byte(0x04)
	date := []byte("11111111")
	service := []byte("ses")
	terminal := []byte("aws4_request")
	message := []byte("SendRawEmail")

	rawSig, err := hmacSignature([]byte("AWS4"+*key), []byte(date))
	if err != nil {
		return "", err
	}

	if rawSig, err = hmacSignature(rawSig, []byte(region)); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, service); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, terminal); err != nil {
		return "", err
	}
	if rawSig, err = hmacSignature(rawSig, message); err != nil {
		return "", err
	}

	versionedSig := make([]byte, 0, len(rawSig)+1)
	versionedSig = append(versionedSig, version)
	versionedSig = append(versionedSig, rawSig...)
	return base64.StdEncoding.EncodeToString(versionedSig), nil
}

func sesSmtpPasswordFromSecretKeyV2(key *string) (string, error) {
	if key == nil {
		return "", nil
	}
	version := byte(0x02)
	message := []byte("SendRawEmail")
	hmacKey := []byte(*key)
	h := hmac.New(sha256.New, hmacKey)
	if _, err := h.Write(message); err != nil {
		return "", err
	}
	rawSig := h.Sum(nil)
	versionedSig := make([]byte, 0, len(rawSig)+1)
	versionedSig = append(versionedSig, version)
	versionedSig = append(versionedSig, rawSig...)
	return base64.StdEncoding.EncodeToString(versionedSig), nil
}
