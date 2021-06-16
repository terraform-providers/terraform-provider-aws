package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSPinpointEventStream_basic(t *testing.T) {
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		ErrorCheck:   testAccErrorCheck(t, pinpoint.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPinpointEventStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEventStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_pinpoint_app.test", "application_id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSPinpointEventStreamConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEventStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_pinpoint_app.test", "application_id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_stream_arn", "aws_kinesis_stream.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSPinpointEventStream_disappears(t *testing.T) {
	var stream pinpoint.EventStream
	resourceName := "aws_pinpoint_event_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		ErrorCheck:   testAccErrorCheck(t, pinpoint.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPinpointEventStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointEventStreamConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEventStreamExists(resourceName, &stream),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsPinpointEventStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccAWSPinpointEventStreamConfig_basic(rName, streamName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_pinpoint_event_stream" "test" {
  application_id         = aws_pinpoint_app.test.application_id
  destination_stream_arn = aws_kinesis_stream.test.arn
  role_arn               = aws_iam_role.test.arn
}

resource "aws_kinesis_stream" "test" {
  name        = %[2]q
  shard_count = 1
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

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
      "${aws_kinesis_stream.test.arn}"
    ]
  }
}
EOF
}
`, rName, streamName)
}

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
