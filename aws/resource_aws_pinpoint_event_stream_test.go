package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/pinpoint"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSPinpointEventStream_basic(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test_event_stream"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointEventStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointEventStreamConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEventStreamExists(resourceName, &stream),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSPinpointEventStreamConfig_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEventStreamExists(resourceName, &stream),
				),
			},
		},
	})
}

func testAccCheckAWSPinpointEventStreamExists(n string, stream *pinpoint.EventStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint event stream with that ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetEventStreamInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetEventStream(params)

		if err != nil {
			return err
		}

		*stream = *output.EventStream

		return nil
	}
}

const testAccAWSPinpointEventStreamConfig_basic = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_event_stream" "test_event_stream" {
  application_id         = "${aws_pinpoint_app.test_app.application_id}"
  destination_stream_arn = "${aws_kinesis_stream.test_stream.arn}"
  role_arn               = "${aws_iam_role.test_role.arn}"
}

resource "aws_kinesis_stream" "test_stream" {
  name        = "terraform-kinesis-test"
  shard_count = 1
}

resource "aws_iam_role" "test_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.us-east-1.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test_role_policy" {
  name   = "test_policy"
  role   = "${aws_iam_role.test_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "kinesis:PutRecords",
      "kinesis:DescribeStream"
    ],
    "Effect": "Allow",
    "Resource": [
      "arn:aws:kinesis:us-east-1:*:*/*"
    ]
  }
}
EOF
}
`

const testAccAWSPinpointEventStreamConfig_update = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_event_stream" "test_event_stream" {
  application_id         = "${aws_pinpoint_app.test_app.application_id}"
  destination_stream_arn = "${aws_kinesis_stream.test_stream_updated.arn}"
  role_arn               = "${aws_iam_role.test_role.arn}"
}

resource "aws_kinesis_stream" "test_stream_updated" {
  name        = "terraform-kinesis-test-updated"
  shard_count = 1
}

resource "aws_iam_role" "test_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.us-east-1.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test_role_policy" {
  name   = "test_policy"
  role   = "${aws_iam_role.test_role.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "kinesis:PutRecords",
      "kinesis:DescribeStream"
    ],
    "Effect": "Allow",
    "Resource": [
      "arn:aws:kinesis:us-east-1:*:*/*"
    ]
  }
}
EOF
}
`

func testAccCheckAWSPinpointEventStreamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_event_stream" {
			continue
		}

		// Check if the event stream exists
		params := &pinpoint.GetEventStreamInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetEventStream(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("Event stream exists when it should be destroyed!")
	}

	return nil
}
