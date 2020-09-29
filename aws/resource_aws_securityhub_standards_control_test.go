package aws

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSecurityHubStandardsControl_basic(t *testing.T) {
	var standardsControl *securityhub.StandardsControl

	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubStandardsControlExists(resourceName, standardsControl),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubStandardsControlConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityHubStandardsControlExists(resourceName, standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_status", "ENABLED"),
					resource.TestMatchResourceAttr(resourceName, "control_status_updated_at", regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)),
					resource.TestCheckResourceAttr(resourceName, "description", "IAM password policies can prevent the reuse of a given password by the same user. It is recommended that the password policy prevent the reuse of passwords."),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", ""),
					resource.TestCheckResourceAttr(resourceName, "related_requirements.0", "CIS AWS Foundations 1.10"),
					resource.TestCheckResourceAttr(resourceName, "severity_rating", "LOW"),
					resource.TestCheckResourceAttr(resourceName, "title", "Ensure IAM password policy prevents password reuse"),
				),
			},
		},
	})
}

func TestAccAWSSecurityHubStandardsControl_disabledControlStatus(t *testing.T) {
	var standardsControl *securityhub.StandardsControl

	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubStandardsControlExists(resourceName, standardsControl),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubStandardsControlConfig_disabledControlStatus(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSecurityHubStandardsControlExists(resourceName, standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", "We handle password policies within Okta"),
				),
			},
		},
	})
}

func TestAccAWSSecurityHubStandardsControl_enabledControlStatusAndDisabledReason(t *testing.T) {
	var standardsControl *securityhub.StandardsControl

	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubStandardsControlExists(resourceName, standardsControl),
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSecurityHubStandardsControlConfig_enabledControlStatus(),
				ExpectError: regexp.MustCompile("InvalidInputException: DisabledReason should not be given for action other than disabling control"),
			},
		},
	})
}

func testAccCheckAWSSecurityHubStandardsControlExists(n string, control *securityhub.StandardsControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		arn := rs.Primary.ID
		subscription_arn := path.Dir(strings.ReplaceAll(arn, "control", "subscription"))

		resp, err := conn.DescribeStandardsControls(&securityhub.DescribeStandardsControlsInput{
			StandardsSubscriptionArn: aws.String(subscription_arn),
		})
		if err != nil {
			return fmt.Errorf("error reading Security Hub %s standard controls: %s", subscription_arn, err)
		}

		controlNotFound := true

		for _, c := range resp.Controls {
			if aws.StringValue(c.StandardsControlArn) != arn {
				continue
			}

			controlNotFound = false
			control = c
		}

		if controlNotFound {
			return fmt.Errorf("Security Hub %s standard control %s not found", subscription_arn, arn)
		}

		return nil
	}
}

func testAccAWSSecurityHubStandardsControlConfig_basic() string {
	return composeConfig(
		testAccAWSSecurityHubStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.10", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
}
`)
}

func testAccAWSSecurityHubStandardsControlConfig_disabledControlStatus() string {
	return composeConfig(
		testAccAWSSecurityHubStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.11", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "DISABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}

func testAccAWSSecurityHubStandardsControlConfig_enabledControlStatus() string {
	return composeConfig(
		testAccAWSSecurityHubStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.12", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}
