package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_config_delivery_channel", &resource.Sweeper{
		Name: "aws_config_delivery_channel",
		Dependencies: []string{
			"aws_config_configuration_recorder",
		},
		F: testSweepConfigDeliveryChannels,
	})
}

func testSweepConfigDeliveryChannels(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).configconn

	req := &configservice.DescribeDeliveryChannelsInput{}
	var resp *configservice.DescribeDeliveryChannelsOutput
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.DescribeDeliveryChannels(req)
		if err != nil {
			// ThrottlingException: Rate exceeded
			if isAWSErr(err, "ThrottlingException", "Rate exceeded") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Delivery Channels sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Delivery Channels: %s", err)
	}

	if len(resp.DeliveryChannels) == 0 {
		log.Print("[DEBUG] No AWS Config Delivery Channel to sweep")
		return nil
	}

	for _, dc := range resp.DeliveryChannels {
		_, err := conn.DeleteDeliveryChannel(&configservice.DeleteDeliveryChannelInput{
			DeliveryChannelName: dc.Name,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Delivery Channel (%s): %s",
				*dc.Name, err)
		}
	}

	return nil
}

func testAccConfigDeliveryChannel_basic(t *testing.T) {
	var dc configservice.DeliveryChannel
	rInt := acctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)
	expectedBucketName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigDeliveryChannelConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigDeliveryChannelExists("aws_config_delivery_channel.foo", &dc),
					testAccCheckConfigDeliveryChannelName("aws_config_delivery_channel.foo", expectedName, &dc),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "name", expectedName),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "s3_bucket_name", expectedBucketName),
				),
			},
		},
	})
}

func testAccConfigDeliveryChannel_allParams(t *testing.T) {
	var dc configservice.DeliveryChannel
	rInt := acctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)
	expectedBucketName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)
	expectedSnsTopicArn := regexp.MustCompile(fmt.Sprintf("arn:aws:sns:[a-z0-9-]+:[0-9]{12}:tf-acc-test-%d", rInt))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigDeliveryChannelConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigDeliveryChannelExists("aws_config_delivery_channel.foo", &dc),
					testAccCheckConfigDeliveryChannelName("aws_config_delivery_channel.foo", expectedName, &dc),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "name", expectedName),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "s3_bucket_name", expectedBucketName),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "s3_key_prefix", "one/two/three"),
					resource.TestMatchResourceAttr("aws_config_delivery_channel.foo", "sns_topic_arn", expectedSnsTopicArn),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "snapshot_delivery_properties.0.delivery_frequency", "Six_Hours"),
				),
			},
		},
	})
}

func testAccConfigDeliveryChannel_importBasic(t *testing.T) {
	resourceName := "aws_config_delivery_channel.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigDeliveryChannelConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConfigDeliveryChannelName(n, desired string, obj *configservice.DeliveryChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if *obj.Name != desired {
			return fmt.Errorf("Expected name: %q, given: %q", desired, *obj.Name)
		}
		return nil
	}
}

func testAccCheckConfigDeliveryChannelExists(n string, obj *configservice.DeliveryChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No delivery channel ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn
		out, err := conn.DescribeDeliveryChannels(&configservice.DescribeDeliveryChannelsInput{
			DeliveryChannelNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe delivery channel: %s", err)
		}
		if len(out.DeliveryChannels) < 1 {
			return fmt.Errorf("No delivery channel found when describing %q", rs.Primary.Attributes["name"])
		}

		dc := out.DeliveryChannels[0]
		*obj = *dc

		return nil
	}
}

func testAccCheckConfigDeliveryChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_delivery_channel" {
			continue
		}

		resp, err := conn.DescribeDeliveryChannels(&configservice.DescribeDeliveryChannelsInput{
			DeliveryChannelNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.DeliveryChannels) != 0 &&
				*resp.DeliveryChannels[0].Name == rs.Primary.Attributes["name"] {
				return fmt.Errorf("Delivery Channel still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigDeliveryChannelConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name = "tf-acc-test-%d"
  role_arn = "${aws_iam_role.r.arn}"
}

resource "aws_iam_role" "r" {
    name = "tf-acc-test-awsconfig-%d"
    assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
    name = "tf-acc-test-awsconfig-%d"
    role = "${aws_iam_role.r.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = "${aws_s3_bucket.b.bucket}"
  depends_on     = ["aws_config_configuration_recorder.foo"]
}`, randInt, randInt, randInt, randInt, randInt)
}

func testAccConfigDeliveryChannelConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name = "tf-acc-test-%d"
  role_arn = "${aws_iam_role.r.arn}"
}

resource "aws_iam_role" "r" {
    name = "tf-acc-test-awsconfig-%d"
    assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
    name = "tf-acc-test-awsconfig-%d"
    role = "${aws_iam_role.r.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_sns_topic" "t" {
  name = "tf-acc-test-%d"
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = "${aws_s3_bucket.b.bucket}"
  s3_key_prefix  = "one/two/three"
  sns_topic_arn  = "${aws_sns_topic.t.arn}"
  snapshot_delivery_properties {
  	delivery_frequency = "Six_Hours"
  }
  depends_on     = ["aws_config_configuration_recorder.foo"]
}`, randInt, randInt, randInt, randInt, randInt, randInt)
}
