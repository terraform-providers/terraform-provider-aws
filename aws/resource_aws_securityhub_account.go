package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSecurityHubAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubAccountCreate,
		Read:   resourceAwsSecurityHubAccountRead,
		Delete: resourceAwsSecurityHubAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsSecurityHubAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Enabling Security Hub for account")

	_, err := conn.EnableSecurityHub(&securityhub.EnableSecurityHubInput{})

	if err != nil {
		return fmt.Errorf("Error enabling Security Hub for account: %s", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceAwsSecurityHubAccountRead(d, meta)
}

func resourceAwsSecurityHubAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Checking if Security Hub is enabled")
	_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

	if err != nil {
		// Can only read enabled standards if Security Hub is enabled
		if isAWSErr(err, "InvalidAccessException", "not subscribed to AWS Security Hub") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error checking if Security Hub is enabled: %s", err)
	}

	return nil
}

func resourceAwsSecurityHubAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disabling Security Hub for account")

	err := resource.Retry(waiter.AdminAccountNotFoundTimeout, func() *resource.RetryError {
		_, err := conn.DisableSecurityHub(&securityhub.DisableSecurityHubInput{})

		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidInputException, "Cannot disable Security Hub on the Security Hub administrator") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DisableSecurityHub(&securityhub.DisableSecurityHubInput{})
	}

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error disabling Security Hub for account: %w", err)
	}

	return nil
}
