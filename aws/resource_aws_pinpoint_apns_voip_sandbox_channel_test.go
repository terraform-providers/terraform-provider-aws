package aws

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/pinpoint"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

/**
 Before running this test, one of the following two ENV variables set must be defined. See here for details:
 https://docs.aws.amazon.com/pinpoint/latest/userguide/channels-mobile-manage.html

 * Key Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_token-based_connection_to_apns )
 APNS_VOIP_BUNDLE_ID    - APNs Bundle ID
 APNS_VOIP_TEAM_ID      - APNs Team ID
 APNS_VOIP_TOKEN_KEY    - Token key file content (.p8 file)
 APNS_VOIP_TOKEN_KEY_ID - APNs Token Key ID

 * Certificate Configuration (ref. https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/establishing_a_certificate-based_connection_to_apns - Select "VoIP Services Certificate" )
 APNS_VOIP_CERTIFICATE             - APNs Certificate content (.pem file content)
 APNS_VOIP_CERTIFICATE_PRIVATE_KEY - APNs Certificate Private Key File content
**/

type testAccAwsPinpointAPNSVoipSandboxChannelCertConfiguration struct {
	Certificate string
	PrivateKey  string
}

type testAccAwsPinpointAPNSVoipSandboxChannelTokenConfiguration struct {
	BundleId   string
	TeamId     string
	TokenKey   string
	TokenKeyId string
}

func testAccAwsPinpointAPNSVoipSandboxChannelCertConfigurationFromEnv(t *testing.T) *testAccAwsPinpointAPNSVoipSandboxChannelCertConfiguration {
	var conf *testAccAwsPinpointAPNSVoipSandboxChannelCertConfiguration
	if os.Getenv("APNS_VOIP_CERTIFICATE") != "" {
		if os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY") == "" {
			t.Fatalf("APNS_VOIP_CERTIFICATE set but missing APNS_VOIP_CERTIFICATE_PRIVATE_KEY")
		}

		conf = &testAccAwsPinpointAPNSVoipSandboxChannelCertConfiguration{
			Certificate: fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE"))),
			PrivateKey:  fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_CERTIFICATE_PRIVATE_KEY"))),
		}
	}

	if conf == nil {
		t.Skipf("Pinpoint certificate credentials envs are missing, skipping test")
	}

	return conf
}

func testAccAwsPinpointAPNSVoipSandboxChannelTokenConfigurationFromEnv(t *testing.T) *testAccAwsPinpointAPNSVoipSandboxChannelTokenConfiguration {
	if os.Getenv("APNS_VOIP_BUNDLE_ID") == "" {
		t.Skipf("APNS_VOIP_BUNDLE_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TEAM_ID") == "" {
		t.Skipf("APNS_VOIP_TEAM_ID env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TOKEN_KEY") == "" {
		t.Skipf("APNS_VOIP_TOKEN_KEY env is missing, skipping test")
	}

	if os.Getenv("APNS_VOIP_TOKEN_KEY_ID") == "" {
		t.Skipf("APNS_VOIP_TOKEN_KEY_ID env is missing, skipping test")
	}

	conf := testAccAwsPinpointAPNSVoipSandboxChannelTokenConfiguration{
		BundleId:   strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_BUNDLE_ID"))),
		TeamId:     strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TEAM_ID"))),
		TokenKey:   fmt.Sprintf("<<EOF\n%s\nEOF\n", strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY"))),
		TokenKeyId: strconv.Quote(strings.TrimSpace(os.Getenv("APNS_VOIP_TOKEN_KEY_ID"))),
	}

	return &conf
}

func TestAccAWSPinpointAPNSVoipSandboxChannel_basicCertificate(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var channel pinpoint.APNSVoipSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_voip_sandbox_channel.test_channel"

	configuration := testAccAwsPinpointAPNSVoipSandboxChannelCertConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAPNSVoipSandboxChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAPNSVoipSandboxChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate", "private_key"},
			},
			{
				Config: testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicCertificate(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAPNSVoipSandboxChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func TestAccAWSPinpointAPNSVoipSandboxChannel_basicToken(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var channel pinpoint.APNSVoipSandboxChannelResponse
	resourceName := "aws_pinpoint_apns_voip_sandbox_channel.test_channel"

	configuration := testAccAwsPinpointAPNSVoipSandboxChannelTokenConfigurationFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAPNSVoipSandboxChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAPNSVoipSandboxChannelExists(resourceName, &channel),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"team_id", "bundle_id", "token_key", "token_key_id"},
			},
			{
				Config: testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicToken(configuration),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAPNSVoipSandboxChannelExists(resourceName, &channel),
				),
			},
		},
	})
}

func testAccCheckAWSPinpointAPNSVoipSandboxChannelExists(n string, channel *pinpoint.APNSVoipSandboxChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint APNs Voip Sandbox Channel with that Application ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetApnsVoipSandboxChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApnsVoipSandboxChannel(params)

		if err != nil {
			return err
		}

		*channel = *output.APNSVoipSandboxChannelResponse

		return nil
	}
}

func testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicCertificate(conf *testAccAwsPinpointAPNSVoipSandboxChannelCertConfiguration) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_sandbox_channel" "test_channel" {
  application_id                = "${aws_pinpoint_app.test_app.application_id}"
  enabled                       = false
  default_authentication_method = "CERTIFICATE"
  certificate                   = %s
  private_key                   = %s
}`, conf.Certificate, conf.PrivateKey)
}

func testAccAWSPinpointAPNSVoipSandboxChannelConfig_basicToken(conf *testAccAwsPinpointAPNSVoipSandboxChannelTokenConfiguration) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}

resource "aws_pinpoint_apns_voip_sandbox_channel" "test_channel" {
  application_id = "${aws_pinpoint_app.test_app.application_id}"
  enabled        = false
  
  default_authentication_method = "TOKEN"

  bundle_id      = %s
  team_id        = %s
  token_key      = %s
  token_key_id   = %s
}`, conf.BundleId, conf.TeamId, conf.TokenKey, conf.TokenKeyId)
}

func testAccCheckAWSPinpointAPNSVoipSandboxChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_apns_voip_sandbox_channel" {
			continue
		}

		// Check if the channel exists
		params := &pinpoint.GetApnsVoipSandboxChannelInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApnsVoipSandboxChannel(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("APNs Voip Sandbox Channel exists when it should be destroyed!")
	}

	return nil
}
