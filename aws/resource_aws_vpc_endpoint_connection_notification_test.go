package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSVpcEndpointConnectionNotification_importBasic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_endpoint_connection_notification.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointConnectionNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConnectionNotificationBasicConfig(lbName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointConnectionNotification_basic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint_connection_notification.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointConnectionNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConnectionNotificationBasicConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointConnectionNotificationExists("aws_vpc_endpoint_connection_notification.foo"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "connection_events.#", "2"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "state", "Enabled"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "notification_type", "Topic"),
				),
			},
			{
				Config: testAccVpcEndpointConnectionNotificationModifiedConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointConnectionNotificationExists("aws_vpc_endpoint_connection_notification.foo"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "connection_events.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "state", "Enabled"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_connection_notification.foo", "notification_type", "Topic"),
				),
			},
		},
	})
}

func testAccCheckVpcEndpointConnectionNotificationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_connection_notification" {
			continue
		}

		resp, err := conn.DescribeVpcEndpointConnectionNotifications(&ec2.DescribeVpcEndpointConnectionNotificationsInput{
			ConnectionNotificationId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidConnectionNotification" {
				continue
			}
			return err
		}
		if len(resp.ConnectionNotificationSet) > 0 {
			return fmt.Errorf("VPC Endpoint connection notification still exist.")
		}

		return err
	}

	return nil
}

func testAccCheckVpcEndpointConnectionNotificationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint connection notification ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeVpcEndpointConnectionNotifications(&ec2.DescribeVpcEndpointConnectionNotificationsInput{
			ConnectionNotificationId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		if len(resp.ConnectionNotificationSet) == 0 {
			return fmt.Errorf("VPC Endpoint connection notification not found")
		}

		return nil
	}
}

func testAccVpcEndpointConnectionNotificationBasicConfig(lbName string) string {
	return fmt.Sprintf(
		`
resource "aws_vpc" "nlb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-connection-notification"
  }
}

resource "aws_lb" "nlb_test" {
  name = "%s"

  subnets = [
    "${aws_subnet.nlb_test_1.id}",
    "${aws_subnet.nlb_test_2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = "testAccVpcEndpointConnectionNotificationBasicConfig_nlb"
  }
}

resource "aws_subnet" "nlb_test_1" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-vpc-endpoint-connection-notification-1"
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-vpc-endpoint-connection-notification-2"
  }
}

data "aws_caller_identity" "current" {}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = false

  network_load_balancer_arns = [
    "${aws_lb.nlb_test.id}",
  ]

  allowed_principals = [
    "${data.aws_caller_identity.current.arn}"
  ]
}

resource "aws_sns_topic" "topic" {
  name = "vpce-notification-topic"

  policy = <<POLICY
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "vpce.amazonaws.com"
        },
        "Action": "SNS:Publish",
        "Resource": "arn:aws:sns:*:*:vpce-notification-topic"
    }]
}
POLICY
}

resource "aws_vpc_endpoint_connection_notification" "foo" {
  vpc_endpoint_service_id = "${aws_vpc_endpoint_service.foo.id}"
  connection_notification_arn = "${aws_sns_topic.topic.arn}"
  connection_events = ["Accept", "Reject"]
}
	`, lbName)
}

func testAccVpcEndpointConnectionNotificationModifiedConfig(lbName string) string {
	return fmt.Sprintf(
		`
		resource "aws_vpc" "nlb_test" {
			cidr_block = "10.0.0.0/16"

	tags = {
				Name = "terraform-testacc-vpc-endpoint-connection-notification"
			}
		}

		resource "aws_lb" "nlb_test" {
			name = "%s"

			subnets = [
				"${aws_subnet.nlb_test_1.id}",
				"${aws_subnet.nlb_test_2.id}",
			]

			load_balancer_type         = "network"
			internal                   = true
			idle_timeout               = 60
			enable_deletion_protection = false

	tags = {
				Name = "testAccVpcEndpointConnectionNotificationBasicConfig_nlb"
			}
		}

		resource "aws_subnet" "nlb_test_1" {
			vpc_id            = "${aws_vpc.nlb_test.id}"
			cidr_block        = "10.0.1.0/24"
			availability_zone = "us-west-2a"

	tags = {
				Name = "tf-acc-vpc-endpoint-connection-notification-1"
			}
		}

		resource "aws_subnet" "nlb_test_2" {
			vpc_id            = "${aws_vpc.nlb_test.id}"
			cidr_block        = "10.0.2.0/24"
			availability_zone = "us-west-2b"

	tags = {
				Name = "tf-acc-vpc-endpoint-connection-notification-2"
			}
		}

		data "aws_caller_identity" "current" {}

		resource "aws_vpc_endpoint_service" "foo" {
			acceptance_required = false

			network_load_balancer_arns = [
				"${aws_lb.nlb_test.id}",
			]

			allowed_principals = [
				"${data.aws_caller_identity.current.arn}"
			]
		}

		resource "aws_sns_topic" "topic" {
			name = "vpce-notification-topic"

			policy = <<POLICY
		{
				"Version":"2012-10-17",
				"Statement":[{
						"Effect": "Allow",
						"Principal": {
								"Service": "vpce.amazonaws.com"
						},
						"Action": "SNS:Publish",
						"Resource": "arn:aws:sns:*:*:vpce-notification-topic"
				}]
		}
		POLICY
		}

		resource "aws_vpc_endpoint_connection_notification" "foo" {
			vpc_endpoint_service_id = "${aws_vpc_endpoint_service.foo.id}"
			connection_notification_arn = "${aws_sns_topic.topic.arn}"
			connection_events = ["Accept"]
		}
	`, lbName)
}
