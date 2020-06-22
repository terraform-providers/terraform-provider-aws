package aws

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsIamUserLoginProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamUserLoginProfileCreate,
		Read:   resourceAwsIamUserLoginProfileRead,
		Delete: resourceAwsIamUserLoginProfileDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("encrypted_password", "")
				d.Set("key_fingerprint", "")
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password_reset_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"password_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      20,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(5, 128),
			},

			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_password": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	charLower   = "abcdefghijklmnopqrstuvwxyz"
	charUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charNumbers = "0123456789"
	charSymbols = "!@#$%^&*()_+-=[]{}|'"
)

// generateIAMPassword generates a random password of a given length, matching the
// most restrictive iam password policy.
func generateIAMPassword(length int) (string, error) {
	const charset = charLower + charUpper + charNumbers + charSymbols

	result := make([]byte, length)
	charsetSize := big.NewInt(int64(len(charset)))

	// rather than trying to artificially add specific characters from each
	// class to the password to match the policy, we generate passwords
	// randomly and reject those that don't match.
	//
	// Even in the worst case, this tends to take less than 10 tries to find a
	// matching password. Any sufficiently long password is likely to succeed
	// on the first try
	for n := 0; n < 100000; n++ {
		for i := range result {
			r, err := rand.Int(rand.Reader, charsetSize)
			if err != nil {
				return "", err
			}
			if !r.IsInt64() {
				return "", errors.New("rand.Int() not representable as an Int64")
			}

			result[i] = charset[r.Int64()]
		}

		if !checkIAMPwdPolicy(result) {
			continue
		}

		return string(result), nil
	}

	return "", errors.New("failed to generate acceptable password")
}

// Check the generated password contains all character classes listed in the
// IAM password policy.
func checkIAMPwdPolicy(pass []byte) bool {
	return (bytes.ContainsAny(pass, charLower) &&
		bytes.ContainsAny(pass, charNumbers) &&
		bytes.ContainsAny(pass, charSymbols) &&
		bytes.ContainsAny(pass, charUpper))
}

func resourceAwsIamUserLoginProfileCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	username := d.Get("user").(string)

	encryptionKey, err := encryption.RetrieveGPGKey(strings.TrimSpace(d.Get("pgp_key").(string)))
	if err != nil {
		return fmt.Errorf("error retrieving GPG Key during IAM User Login Profile (%s) creation: %s", username, err)
	}

	passwordResetRequired := d.Get("password_reset_required").(bool)
	passwordLength := d.Get("password_length").(int)
	initialPassword, err := generateIAMPassword(passwordLength)
	if err != nil {
		return err
	}

	fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, initialPassword, "Password")
	if err != nil {
		return fmt.Errorf("error encrypting password during IAM User Login Profile (%s) creation: %s", username, err)
	}

	request := &iam.CreateLoginProfileInput{
		UserName:              aws.String(username),
		Password:              aws.String(initialPassword),
		PasswordResetRequired: aws.Bool(passwordResetRequired),
	}

	log.Println("[DEBUG] Create IAM User Login Profile request:", request)
	createResp, err := iamconn.CreateLoginProfile(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM User Login Profile for %q: %s", username, err)
	}

	d.SetId(*createResp.LoginProfile.UserName)
	d.Set("key_fingerprint", fingerprint)
	d.Set("encrypted_password", encrypted)
	return nil
}

func resourceAwsIamUserLoginProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.GetLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Getting IAM User Login Profile (%s): %s", d.Id(), input)
	output, err := conn.GetLoginProfile(input)

	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM User Login Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting IAM User Login Profile (%s): %s", d.Id(), err)
	}

	if output == nil || output.LoginProfile == nil {
		return fmt.Errorf("error getting IAM User Login Profile (%s): empty response", d.Id())
	}

	d.Set("user", aws.StringValue(output.LoginProfile.UserName))

	return nil
}

func resourceAwsIamUserLoginProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.DeleteLoginProfileInput{
		UserName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting IAM User Login Profile (%s): %s", d.Id(), input)
	// Handle IAM eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteLoginProfile(input)

		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}

		// EntityTemporarilyUnmodifiable: Login Profile for User XXX cannot be modified while login profile is being created.
		if isAWSErr(err, iam.ErrCodeEntityTemporarilyUnmodifiableException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Handle AWS Go SDK automatic retries
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteLoginProfile(input)
	}

	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IAM User Login Profile (%s): %s", d.Id(), err)
	}

	return nil
}
