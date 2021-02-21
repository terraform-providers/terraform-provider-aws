package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSpotDatafeedSubscription_serial(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"basic":      testAccAWSSpotDatafeedSubscription_basic,
		"disappears": testAccAWSSpotDatafeedSubscription_disappears,
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSSpotDatafeedSubscription_basic(t *testing.T) {
	var subscription ec2.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSpotDatafeedSubscription(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotDatafeedSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotDatafeedSubscription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotDatafeedSubscriptionExists(resourceName, &subscription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSSpotDatafeedSubscriptionDisappears(subscription *ec2.SpotDatafeedSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DeleteSpotDatafeedSubscription(&ec2.DeleteSpotDatafeedSubscriptionInput{})
		if err != nil {
			return err
		}

		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			_, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok && cgw.Code() == "InvalidSpotDatafeed.NotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Spot Datafeed Subscription: %s", err))
			}
			return resource.RetryableError(fmt.Errorf("Waiting for Spot Datafeed Subscription"))
		})
	}
}

func testAccAWSSpotDatafeedSubscription_disappears(t *testing.T) {
	var subscription ec2.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSpotDatafeedSubscription(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotDatafeedSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotDatafeedSubscription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSpotDatafeedSubscriptionExists(resourceName, &subscription),
					testAccCheckAWSSpotDatafeedSubscriptionDisappears(&subscription),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSpotDatafeedSubscriptionExists(n string, subscription *ec2.SpotDatafeedSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})
		if err != nil {
			return err
		}

		*subscription = *resp.SpotDatafeedSubscription

		return nil
	}
}

func testAccCheckAWSSpotDatafeedSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_datafeed_subscription" {
			continue
		}

		_, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})

		if tfawserr.ErrCodeEquals(err, "InvalidSpotDatafeed.NotFound") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error descripting EC2 Spot Datafeed Subscription: %w", err)
		}
	}

	return nil
}

func testAccPreCheckAWSSpotDatafeedSubscription(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeSpotDatafeedSubscriptionInput{}

	_, err := conn.DescribeSpotDatafeedSubscription(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfawserr.ErrCodeEquals(err, "InvalidSpotDatafeed.NotFound") {
		return
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSSpotDatafeedSubscription(rName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }

  grant {
    id          = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" # EC2 Account
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
}

resource "aws_spot_datafeed_subscription" "test" {
  bucket = aws_s3_bucket.test.bucket
}
`, rName)
}
