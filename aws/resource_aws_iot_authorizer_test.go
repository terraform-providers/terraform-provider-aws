package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iot"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIoTAuthorizer_basic(t *testing.T) {
	var conf iot.DescribeAuthorizerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
				),
			},
			{
				Config: testAccAWSIoTAuthorizerConfig_defaultStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", iot.AuthorizerStatusActive),
				),
			},
			{
				Config: testAccAWSIoTAuthorizerConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signing_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", iot.AuthorizerStatusInactive),
				),
			},
		},
	})
}

func TestAccAWSIoTAuthorizer_disappears(t *testing.T) {
	var conf iot.DescribeAuthorizerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_authorizer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAuthorizerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAuthorizerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAuthorizerExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIoTAuthorizer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSIoTAuthorizerExists(n string, res *iot.DescribeAuthorizerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Authorizer ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn

		req := &iot.DescribeAuthorizerInput{
			AuthorizerName: aws.String(rs.Primary.ID),
		}
		describe, err := conn.DescribeAuthorizer(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSIoTAuthorizerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_authorizer" {
			continue
		}

		req := &iot.DescribeAuthorizerInput{
			AuthorizerName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeAuthorizer(req)

		if err == nil {
			return fmt.Errorf("IoT Authorizer still exists")
		}

		aws2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if aws2err.Code() != iot.ErrCodeResourceNotFoundException {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSIoTAuthorizerConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

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

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}
`, rName)
}

func testAccAWSIoTAuthorizerConfig(rName string) string {
	return testAccAWSIoTAuthorizerConfigBase(rName) + fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                      = %[1]q
  authorizer_function_arn   = aws_lambda_function.test.arn
  signing_disabled          = false
  status                    = "ACTIVE"
  token_key_name            = "Token-Header"
  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authroizer-signing-key.pem")}"
  }
}
`, rName)
}

func testAccAWSIoTAuthorizerConfig_update(rName string) string {
	return testAccAWSIoTAuthorizerConfigBase(rName) + fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                      = %[1]q
  authorizer_function_arn   = aws_lambda_function.test.arn
  signing_disabled          = false
  token_key_name            = "Token-Header"
  status                    = "INACTIVE"
  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authroizer-signing-key.pem")}"
    Key2 = "${file("test-fixtures/iot-authroizer-signing-key.pem")}"
  }
}
`, rName)
}

func testAccAWSIoTAuthorizerConfig_defaultStatus(rName string) string {
	return testAccAWSIoTAuthorizerConfigBase(rName) + fmt.Sprintf(`
resource "aws_iot_authorizer" "test" {
  name                      = %[1]q
  authorizer_function_arn   = aws_lambda_function.test.arn
  signing_disabled          = false
  token_key_name            = "Token-Header"
  token_signing_public_keys = {
    Key1 = "${file("test-fixtures/iot-authroizer-signing-key.pem")}"
  }
}
`, rName)
}
