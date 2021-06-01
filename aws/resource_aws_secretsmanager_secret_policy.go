package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

func resourceAwsSecretsManagerSecretPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecretsManagerSecretPolicyCreate,
		Read:   resourceAwsSecretsManagerSecretPolicyRead,
		Update: resourceAwsSecretsManagerSecretPolicyUpdate,
		Delete: resourceAwsSecretsManagerSecretPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"secret_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceAwsSecretsManagerSecretPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	input := &secretsmanager.PutResourcePolicyInput{
		ResourcePolicy: aws.String(d.Get("policy").(string)),
		SecretId:       aws.String(d.Get("secret_arn").(string)),
	}

	if v, ok := d.GetOk("block_public_policy"); ok {
		input.BlockPublicPolicy = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %#v", input)
	var res *secretsmanager.PutResourcePolicyOutput

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		res, err = conn.PutResourcePolicy(input)
		if isAWSErr(err, secretsmanager.ErrCodeMalformedPolicyDocumentException,
			"This resource policy contains an unsupported principal") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		res, err = conn.PutResourcePolicy(input)
	}
	if err != nil {
		return fmt.Errorf("error setting Secrets Manager Secret %q policy: %w", d.Id(), err)
	}

	d.SetId(aws.StringValue(res.ARN))

	return resourceAwsSecretsManagerSecretPolicyRead(d, meta)
}

func resourceAwsSecretsManagerSecretPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	input := &secretsmanager.GetResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Secrets Manager Secret Policy: %#v", input)
	res, err := conn.GetResourcePolicy(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] SecretsManager Secret Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Secrets Manager Secret policy: %w", err)
	}

	if res.ResourcePolicy != nil {
		policy, err := structure.NormalizeJsonString(aws.StringValue(res.ResourcePolicy))
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %w", err)
		}
		d.Set("policy", policy)
	} else {
		d.Set("policy", "")
	}
	d.Set("secret_arn", d.Id())

	return nil
}

func resourceAwsSecretsManagerSecretPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	if d.HasChanges("policy", "block_public_policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		input := &secretsmanager.PutResourcePolicyInput{
			ResourcePolicy:    aws.String(policy),
			SecretId:          aws.String(d.Id()),
			BlockPublicPolicy: aws.Bool(d.Get("block_public_policy").(bool)),
		}

		log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %#v", input)
		err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			var err error
			_, err = conn.PutResourcePolicy(input)
			if isAWSErr(err, secretsmanager.ErrCodeMalformedPolicyDocumentException,
				"This resource policy contains an unsupported principal") {
				return resource.RetryableError(err)
			}
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.PutResourcePolicy(input)
		}
		if err != nil {
			return fmt.Errorf("error setting Secrets Manager Secret %q policy: %w", d.Id(), err)
		}
	}

	return resourceAwsSecretsManagerSecretPolicyRead(d, meta)
}

func resourceAwsSecretsManagerSecretPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	input := &secretsmanager.DeleteResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Removing Secrets Manager Secret policy: %#v", input)
	_, err := conn.DeleteResourcePolicy(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error removing Secrets Manager Secret %q policy: %w", d.Id(), err)
	}

	return nil
}
