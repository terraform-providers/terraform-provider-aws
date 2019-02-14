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

/**
 Before running this test, the following ENV variable must be set:

 GCM_API_KEY - Google Cloud Messaging Api Key
**/

func TestAccAWSPinpointGCMChannel_basic(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var channel pinpoint.GCMChannelResponse
	resourceName := "aws_pinpoint_gcm_channel.test_gcm_channel"

	if os.Getenv("GCM_API_KEY") == "" {
		t.Skipf("GCM_API_KEY env missing, skip test")
	}

	apiKey := os.Getenv("GCM_API_KEY")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointGCMChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointGCMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"api_key"},
			},
			{
				Config: testAccAWSPinpointGCMChannelConfig_basic(apiKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointGCMChannelExists(resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSPinpointGCMChannelExists(n string, channel *pinpoint.GCMChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint GCM Channel with that application ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetGcmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetGcmChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.GCMChannelResponse

		return nil
	}
}

func testAccAWSPinpointGCMChannelConfig_basic(apiKey string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_gcm_channel" "test_gcm_channel" {
  application_id = "${aws_pinpoint_app.test_app.application_id}"
  enabled        = "false"
  api_key        = "%s"
}`, apiKey)
}

func testAccCheckAWSPinpointGCMChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_gcm_channel" {
			continue
		}

		// Check if the event stream exists
		params := &pinpoint.GetGcmChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetGcmChannel(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("GCM Channel exists when it should be destroyed!")
	}

	return nil
}
