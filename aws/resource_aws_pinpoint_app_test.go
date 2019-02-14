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

func TestAccAWSPinpointApp_basic(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test_app"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_withGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
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

func TestAccAWSPinpointApp_CampaignHookLambda(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test_app"
	appName := "terraform-test-pinpointapp-campaignhooklambda"
	lambdaName := "test-pinpoint-lambda"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_pinpoint_app.test_app",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_CampaignHookLambda(appName, lambdaName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "campaign_hook.0.mode", "DELIVERY"),
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

func TestAccAWSPinpointApp_Limits(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test_app"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_pinpoint_app.test_app",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_Limits,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "limits.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "limits.0.total", "500"),
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

func TestAccAWSPinpointApp_QuietTime(t *testing.T) {
	oldDefaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldDefaultRegion)

	var application pinpoint.ApplicationResponse
	resourceName := "aws_pinpoint_app.test_app"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_pinpoint_app.test_app",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSPinpointAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPinpointAppConfig_QuietTime,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPinpointAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quiet_time.0.start", "00:00"),
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

func testAccCheckAWSPinpointAppExists(n string, application *pinpoint.ApplicationResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint app with that ID exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).pinpointconn

		// Check if the app exists
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetApp(params)

		if err != nil {
			return err
		}

		*application = *output.ApplicationResponse

		return nil
	}
}

const testAccAWSPinpointAppConfig_withGeneratedName = `
provider "aws" {
	region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {}
`

func testAccAWSPinpointAppConfig_CampaignHookLambda(appName, funcName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}


resource "aws_pinpoint_app" "test_app" {
    name = "%s"

    campaign_hook {
        lambda_function_name = "${aws_lambda_function.test.arn}"
        mode                 = "DELIVERY"
    }
}

resource "aws_lambda_function" "test" {
    filename      = "test-fixtures/lambdapinpoint.zip"
    function_name = "%s"
    role          = "${aws_iam_role.iam_for_lambda.arn}"
    handler       = "lambdapinpoint.handler"
    runtime       = "nodejs6.10"
    publish       = true
}

resource "aws_iam_role" "iam_for_lambda" {
    name = "test-role"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_caller_identity" "aws" {}

resource "aws_lambda_permission" "permission" {
  statement_id = "AllowExecutionFromPinpoint"
  action = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.test.function_name}"
  principal = "pinpoint.us-east-1.amazonaws.com"
  source_arn = "arn:aws:mobiletargeting:us-east-1:${data.aws_caller_identity.aws.account_id}:/apps/*"
}
`, appName, funcName)
}

const testAccAWSPinpointAppConfig_Limits = `
provider "aws" {
	region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {
    name = "terraform-test-pinpointapp-limits"

    limits {
        daily               = 3
        maximum_duration    = 600
        messages_per_second = 50
        total               = 500
    }
}
`

const testAccAWSPinpointAppConfig_QuietTime = `
provider "aws" {
	region = "us-east-1"
}

resource "aws_pinpoint_app" "test_app" {
    name = "terraform-test-pinpointapp-quiet"

    quiet_time {
        start = "00:00"
        end   = "03:00"
    }
}
`

func testAccCheckAWSPinpointAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).pinpointconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpoint_app" {
			continue
		}

		// Check if the topic exists by fetching its attributes
		params := &pinpoint.GetAppInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetApp(params)
		if err != nil {
			if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("App exists when it should be destroyed!")
	}

	return nil
}
