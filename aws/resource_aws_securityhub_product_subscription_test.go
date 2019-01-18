package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/securityhub"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAWSSecurityHubProductSubscription_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				// We would like to use an AWS product subscription, but they are
				// all automatically subscribed when enabling Security Hub.
				// This configuration will enable Security Hub, then in a later PreConfig,
				// we will disable an AWS product subscription so we can test (re-)enabling it.
				Config: testAccAWSSecurityHubProductSubscriptionConfig_empty,
				Check:  testAccCheckAWSSecurityHubAccountExists("aws_securityhub_account.example"),
			},
			{
				// AWS product subscriptions happen automatically when enabling Security Hub.
				// Here we attempt to remove one so we can attempt to (re-)enable it.
				PreConfig: func() {
					conn := testAccProvider.Meta().(*AWSClient).securityhubconn
					productSubscriptionARN := arn.ARN{
						AccountID: testAccGetAccountID(),
						Partition: testAccGetPartition(),
						Region:    testAccGetRegion(),
						Resource:  "product-subscription/aws/guardduty",
						Service:   "securityhub",
					}.String()

					input := &securityhub.DisableImportFindingsForProductInput{
						ProductSubscriptionArn: aws.String(productSubscriptionARN),
					}

					_, err := conn.DisableImportFindingsForProduct(input)
					if err != nil {
						t.Fatalf("error disabling Security Hub Product Subscription for GuardDuty: %s", err)
					}
				},
				Config: testAccAWSSecurityHubProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubProductSubscriptionExists("aws_securityhub_product_subscription.example"),
				),
			},
			{
				ResourceName:      "aws_securityhub_product_subscription.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccAWSSecurityHubProductSubscriptionConfig_empty,
				Check:  testAccCheckAWSSecurityHubProductSubscriptionDestroy,
			},
		},
	})
}

func testAccCheckAWSSecurityHubProductSubscriptionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		_, productSubscriptionArn, err := resourceAwsSecurityHubProductSubscriptionParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, err := resourceAwsSecurityHubProductSubscriptionCheckExists(conn, productSubscriptionArn)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Security Hub product subscription %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSSecurityHubProductSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_product_subscription" {
			continue
		}

		_, productSubscriptionArn, err := resourceAwsSecurityHubProductSubscriptionParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, err := resourceAwsSecurityHubProductSubscriptionCheckExists(conn, productSubscriptionArn)

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Security Hub product subscription %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccAWSSecurityHubProductSubscriptionConfig_empty = `
resource "aws_securityhub_account" "example" {}
`

const testAccAWSSecurityHubProductSubscriptionConfig_basic = `
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = ["aws_securityhub_account.example"]
  product_arn = "arn:aws:securityhub:${data.aws_region.current.name}::product/aws/guardduty"
}
`
