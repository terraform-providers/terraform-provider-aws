package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsSecurityHubAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubAccountCreate,
		Read:   resourceAwsSecurityHubAccountRead,
		Update: resourceAwsSecurityHubAccountUpdate,
		Delete: resourceAwsSecurityHubAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"enable_default_standards": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceAwsSecurityHubAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.EnableSecurityHubInput{
		EnableDefaultStandards: aws.Bool(d.Get("enable_default_standards").(bool)),
	}

	_, err := conn.EnableSecurityHub(input)

	if err != nil {
		return fmt.Errorf("error enabling Security Hub for account: %w", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	if aws.BoolValue(input.EnableDefaultStandards) {
		if err := waiter.StandardsSubscriptionsReady(conn); err != nil {
			return fmt.Errorf("error waiting for Security Hub default standards to be enabled for account: %w", err)
		}
	}

	return resourceAwsSecurityHubAccountRead(d, meta)
}

func resourceAwsSecurityHubAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	standardsSubscriptions, err := finder.EnabledStandardsSubscriptions(conn)

	// Can only read enabled standards if Security Hub is enabled
	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
		log.Printf("[WARN] Securty Hub for account not found, removing from state")
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error checking if Security Hub for account is enabled: %w", err)
	}

	if len(standardsSubscriptions) == 0 {
		d.Set("enable_default_standards", false)
		return nil
	}

	standards, err := standardsEnabledByDefault(conn)

	if err != nil {
		return fmt.Errorf("error describing Security Hub standards enabled by default for account: %w", err)
	}

	d.Set("enable_default_standards", defaultStandardsEnabled(standards, standardsSubscriptions))

	return nil
}

func resourceAwsSecurityHubAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	if d.HasChange("enable_default_standards") {
		enable := d.Get("enable_default_standards").(bool)

		standards, err := standardsEnabledByDefault(conn)

		if err != nil {
			return fmt.Errorf("error describing Security Hub standards enabled by default for account: %w", err)
		}

		if enable {
			requests := make([]*securityhub.StandardsSubscriptionRequest, len(standards))

			for i, standard := range standards {
				request := &securityhub.StandardsSubscriptionRequest{
					StandardsArn: standard.StandardsArn,
				}

				requests[i] = request
			}

			_, err = conn.BatchEnableStandards(&securityhub.BatchEnableStandardsInput{
				StandardsSubscriptionRequests: requests,
			})

			if err != nil {
				return fmt.Errorf("error enabling Security Hub default standards for account: %w", err)
			}

			if err := waiter.StandardsSubscriptionsReady(conn); err != nil {
				return fmt.Errorf("error waiting for Security Hub default standards to be enabled for account: %w", err)
			}
		} else {
			standardsSubscriptions, err := standardsSubscriptionsEnabledByDefault(conn, standards)

			if err != nil {
				return fmt.Errorf("error getting Security Hub standards subscriptions enabled by default: %w", err)
			}

			subscriptionArns := make([]*string, len(standardsSubscriptions))

			for i, subscription := range standardsSubscriptions {
				if subscription == nil {
					continue
				}
				subscriptionArns[i] = subscription.StandardsSubscriptionArn
			}

			_, err = conn.BatchDisableStandards(&securityhub.BatchDisableStandardsInput{
				StandardsSubscriptionArns: subscriptionArns,
			})

			if err != nil {
				return fmt.Errorf("error disabling Security Hub default standards for account: %w", err)
			}

			if err := waiter.StandardsSubscriptionsDeleted(conn); err != nil {
				return fmt.Errorf("error waiting for Security Hub default standards to be disabled for account: %w", err)
			}
		}
	}

	return nil
}

func resourceAwsSecurityHubAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

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
		return fmt.Errorf("error disabling Security Hub for account: %w", err)
	}

	return nil
}

// defaultStandardsEnabled returns true if the list of Standards Subscriptions
// correspond to the Standards Security Hub enables by default; otherwise, returns false
func defaultStandardsEnabled(standards []*securityhub.Standard, subscriptions []*securityhub.StandardsSubscription) bool {
	defaults := schema.NewSet(schema.HashString, nil)
	enabled := schema.NewSet(schema.HashString, nil)

	for _, standard := range standards {
		if standard == nil {
			continue
		}

		defaults.Add(aws.StringValue(standard.StandardsArn))
	}

	for _, subscription := range subscriptions {
		if subscription == nil {
			continue
		}

		enabled.Add(aws.StringValue(subscription.StandardsArn))
	}

	return enabled.Intersection(defaults).Equal(defaults)
}

// standardsEnabledByDefault returns a list of the standards that Security Hub enables by default.
func standardsEnabledByDefault(conn *securityhub.SecurityHub) ([]*securityhub.Standard, error) {
	var defaults []*securityhub.Standard

	err := conn.DescribeStandardsPages(&securityhub.DescribeStandardsInput{}, func(page *securityhub.DescribeStandardsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, standard := range page.Standards {
			if standard == nil {
				continue
			}
			if aws.BoolValue(standard.EnabledByDefault) {
				defaults = append(defaults, standard)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return defaults, nil
}

// standardsSubscriptionsEnabledByDefault returns a list of the standards subscriptions that Security Hub enables by default.
func standardsSubscriptionsEnabledByDefault(conn *securityhub.SecurityHub, standards []*securityhub.Standard) ([]*securityhub.StandardsSubscription, error) {
	var standardsSubscriptions []*securityhub.StandardsSubscription
	var errors *multierror.Error

	for _, standard := range standards {
		if standard == nil {
			continue
		}

		enabledStandardsSubscriptions, err := finder.EnabledStandardsSubscriptions(conn)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("error getting Security Hub enabled standards subscription for standard (%s): %w", aws.StringValue(standard.Name), err))
			continue
		}

		for _, enabledStandardsSubscription := range enabledStandardsSubscriptions {
			if aws.StringValue(enabledStandardsSubscription.StandardsArn) == aws.StringValue(standard.StandardsArn) {
				standardsSubscriptions = append(standardsSubscriptions, enabledStandardsSubscription)
			}
		}
	}

	return standardsSubscriptions, errors.ErrorOrNil()
}
