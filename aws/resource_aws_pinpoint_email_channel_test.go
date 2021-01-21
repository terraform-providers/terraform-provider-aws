package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSPinpointEmailChannel_basic(t *testing.T) {
	var channel pinpoint.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test_email_channel"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSPinpointApp(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointEmailChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointEmailChannelConfig_FromAddress("user@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEmailChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "messages_per_second"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSPinpointEmailChannelConfig_FromAddress("userupdated@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointEmailChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "messages_per_second"),
				),
			},
		},
	})
}

func testAccCheckAWSPinpointEmailChannelExists(n string, channel *pinpoint.EmailChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint Email Channel with that application ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetEmailChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetEmailChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.EmailChannelResponse

		return nil
	}
}

func testAccAWSPinpointEmailChannelConfig_FromAddress(fromAddress string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_email_channel" "test_email_channel" {
  application_id = aws_pinpoint_app.test_app.application_id
  enabled        = "false"
  from_address   = %[1]q
  identity       = aws_ses_domain_identity.test_identity.arn
  role_arn       = aws_iam_role.test_role.arn
}

resource "aws_ses_domain_identity" "test_identity" {
  domain = "example.com"
}

resource "aws_iam_role" "test_role" {
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

resource "aws_iam_role_policy" "test_role_policy" {
  name = "test_policy"
  role = aws_iam_role.test_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "mobileanalytics:PutEvents",
      "mobileanalytics:PutItems"
    ],
    "Effect": "Allow",
    "Resource": [
      "*"
    ]
  }
}
EOF
}
`, fromAddress)
}

func testAccCheckAWSPinpointEmailChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_email_channel" {
			continue
		}

		// Check if the event stream exists
		params := &pinpoint.GetEmailChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetEmailChannel(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("Email Channel exists when it should be destroyed!")
	}

	return nil
}
