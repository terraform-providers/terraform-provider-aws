package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsIoTAuthorizer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAuthorizerCreate,
		Read:   resourceAwsIotAuthorizerRead,
		Update: resourceAwsIotAuthorizerUpdate,
		Delete: resourceAwsIotAuthorizerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[\w=,@-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"authorizer_function_arn": {
				Type:     schema.TypeString,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"signing_disabled": {
				Type:     schema.TypeBool,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeBool,
				},
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  iot.AuthorizerStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					iot.AuthorizerStatusActive,
					iot.AuthorizerStatusInactive,
				}, false),
			},
			"token_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"token_signing_public_keys": {
				Type:      schema.TypeMap,
				Required:  true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Sensitive: true,
			},
		},
	}
}

func resourceAwsIotAuthorizerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	tokenSigningPublicKeys := make(map[string]string)
	for k, v := range d.Get("token_signing_public_keys").(map[string]interface{}) {
		tokenSigningPublicKeys[k] = v.(string)
	}

	input := iot.CreateAuthorizerInput{
		AuthorizerFunctionArn:  aws.String(d.Get("authorizer_function_arn").(string)),
		AuthorizerName:         aws.String(d.Get("name").(string)),
		SigningDisabled:        aws.Bool(d.Get("signing_disabled").(bool)),
		Status:                 aws.String(d.Get("status").(string)),
		TokenKeyName:           aws.String(d.Get("token_key_name").(string)),
		TokenSigningPublicKeys: aws.StringMap(tokenSigningPublicKeys),
	}

	log.Printf("[INFO] Creating IoT Authorizer: %s", input)
	out, err := conn.CreateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Error creating IoT Authorizer: %w", err)
	}

	d.SetId(aws.StringValue(out.AuthorizerName))

	return resourceAwsIotAuthorizerRead(d, meta)
}

func resourceAwsIotAuthorizerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	log.Printf("[INFO] Reading IoT Authorizer %s", d.Id())
	input := iot.DescribeAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	}

	authorizerDescription, err := conn.DescribeAuthorizer(&input)
	if err != nil {
		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] No IoT Authorizer (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	authorizer := authorizerDescription.AuthorizerDescription
	log.Printf("[DEBUG] Received IoT Authorizer: %s", authorizer)

	d.Set("authorizer_function_arn", authorizer.AuthorizerFunctionArn)
	d.Set("signing_disabled", authorizer.SigningDisabled)
	d.Set("status", authorizer.Status)
	d.Set("token_key_name", authorizer.TokenKeyName)
	d.Set("name", authorizer.AuthorizerName)
	if err := d.Set("token_signing_public_keys", authorizer.TokenSigningPublicKeys); err != nil {
		return fmt.Errorf("Error setting token signing keys for IoT Authrozer: %s", err)
	}

	return nil
}

func resourceAwsIotAuthorizerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	tokenSigningPublicKeys := make(map[string]string)
	for k, v := range d.Get("token_signing_public_keys").(map[string]interface{}) {
		tokenSigningPublicKeys[k] = v.(string)
	}

	input := iot.UpdateAuthorizerInput{
		AuthorizerFunctionArn:  aws.String(d.Get("authorizer_function_arn").(string)),
		AuthorizerName:         aws.String(d.Id()),
		Status:                 aws.String(d.Get("status").(string)),
		TokenKeyName:           aws.String(d.Get("token_key_name").(string)),
		TokenSigningPublicKeys: aws.StringMap(tokenSigningPublicKeys),
	}

	log.Printf("[INFO] Updating IoT Authorizer: %s", input)
	_, err := conn.UpdateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Updating IoT Authorizer failed: %w", err)
	}

	return resourceAwsIotAuthorizerRead(d, meta)
}

func resourceAwsIotAuthorizerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	input := iot.DeleteAuthorizerInput{
		AuthorizerName: aws.String(d.Id()),
	}

	status := d.Get("status").(string)

	// In order to delete an IoT Authorizer, you must set it inactive first.
	if status == iot.AuthorizerStatusActive {
		setInactiveError := d.Set("status", iot.AuthorizerStatusInactive)

		if setInactiveError != nil {
			return fmt.Errorf("Setting IoT Authorizer to INACTIVE failed: %w", setInactiveError)

		}
		updateErr := resourceAwsIotAuthorizerUpdate(d, meta)

		if updateErr != nil {
			return fmt.Errorf("Updating IoT Authorizer to INACTIVE failed: %w", updateErr)
		}
	}

	log.Printf("[INFO] Deleting IoT Authorizer: %s", input)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteAuthorizer(&input)

		if err != nil {
			if isAWSErr(err, iot.ErrCodeInvalidRequestException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteAuthorizer(&input)
		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting IOT Authorizer: %s", err)
	}
	return nil
}
