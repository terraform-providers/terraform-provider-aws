package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayAccount_basic(t *testing.T) {
	var conf apigateway.Account

	rInt := acctest.RandInt()
	firstName := fmt.Sprintf("tf_acc_api_gateway_cloudwatch_%d", rInt)
	secondName := fmt.Sprintf("tf_acc_api_gateway_cloudwatch_modified_%d", rInt)
	resourceName := "aws_api_gateway_account.test"
	expectedRoleArn_first := regexp.MustCompile("role/" + firstName + "$")
	expectedRoleArn_second := regexp.MustCompile("role/" + secondName + "$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayAccountConfig_updated(firstName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAccountExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayAccountCloudwatchRoleArn(&conf, expectedRoleArn_first),
					testAccMatchResourceAttrGlobalARN(resourceName, "cloudwatch_role_arn", "iam", expectedRoleArn_first),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cloudwatch_role_arn"},
			},
			{
				Config: testAccAWSAPIGatewayAccountConfig_updated2(secondName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAccountExists(resourceName, &conf),
					testAccCheckAWSAPIGatewayAccountCloudwatchRoleArn(&conf, expectedRoleArn_second),
					testAccMatchResourceAttrGlobalARN(resourceName, "cloudwatch_role_arn", "iam", expectedRoleArn_second),
				),
			},
			{
				Config: testAccAWSAPIGatewayAccountConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayAccountExists(resourceName, &conf),
					// This resource does not un-set the value, so this will preserve the CloudWatch role ARN setting on the
					// deployed resource, but will be empty in the Terraform state
					testAccCheckAWSAPIGatewayAccountCloudwatchRoleArn(&conf, expectedRoleArn_second),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_role_arn", ""),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayAccountCloudwatchRoleArn(conf *apigateway.Account, expectedArn *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expectedArn == nil && conf.CloudwatchRoleArn == nil {
			return nil
		}
		if expectedArn == nil && conf.CloudwatchRoleArn != nil {
			return fmt.Errorf("Expected empty CloudwatchRoleArn, given: %q", *conf.CloudwatchRoleArn)
		}
		if expectedArn != nil && conf.CloudwatchRoleArn == nil {
			return fmt.Errorf("Empty CloudwatchRoleArn, expected: %q", expectedArn)
		}
		if !expectedArn.MatchString(*conf.CloudwatchRoleArn) {
			return fmt.Errorf("CloudwatchRoleArn didn't match. Expected: %q, Given: %q", expectedArn, *conf.CloudwatchRoleArn)
		}
		return nil
	}
}

func testAccCheckAWSAPIGatewayAccountExists(n string, res *apigateway.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Account ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

		req := &apigateway.GetAccountInput{}
		describe, err := conn.GetAccount(req)
		if err != nil {
			return err
		}
		if describe == nil {
			return fmt.Errorf("Got nil account ?!")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayAccountDestroy(s *terraform.State) error {
	// Intentionally noop
	// as there is no API method for deleting or resetting account settings
	return nil
}

// testAccPreCheckAWSAPIGatewayAccountCloudWatchRoleArn checks whether a CloudWatch role ARN has been configured in the current AWS region.
func testAccPreCheckAWSAPIGatewayAccountCloudWatchRoleArn(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

	output, err := conn.GetAccount(&apigateway.GetAccountInput{})

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error reading API Gateway Account: %s", err)
	}

	if output == nil || aws.StringValue(output.CloudwatchRoleArn) == "" {
		t.Skip("skipping tests; no API Gateway CloudWatch role ARN has been configured in this region")
	}
}

const testAccAWSAPIGatewayAccountConfig_empty = `
resource "aws_api_gateway_account" "test" {
}
`

func testAccAWSAPIGatewayAccountConfig_updated(randName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}

resource "aws_iam_role" "cloudwatch" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "default"
  role = aws_iam_role.cloudwatch.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}
`, randName)
}

func testAccAWSAPIGatewayAccountConfig_updated2(randName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.second.arn
}

resource "aws_iam_role" "second" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "cloudwatch" {
  name = "default"
  role = aws_iam_role.second.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}
EOF
}
`, randName)
}
