package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLambdaProvisionedConcurrencyConfig_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigQualifierFunctionVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaFunctionResourceName, "version"),
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

func TestAccAWSLambdaProvisionedConcurrencyConfig_disappears_LambdaFunction(t *testing.T) {
	var function lambda.GetFunctionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigProvisionedConcurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(lambdaFunctionResourceName, rName, &function),
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					testAccCheckAwsLambdaFunctionDisappears(&function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaProvisionedConcurrencyConfig_disappears_LambdaProvisionedConcurrencyConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigProvisionedConcurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					testAccCheckAwsLambdaProvisionedConcurrencyConfigDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaProvisionedConcurrencyConfig_ProvisionedConcurrentExecutions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigProvisionedConcurrentExecutions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigProvisionedConcurrentExecutions(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_concurrent_executions", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", "1"),
				),
			},
		},
	})
}

func TestAccAWSLambdaProvisionedConcurrencyConfig_Qualifier_AliasName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_provisioned_concurrency_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaProvisionedConcurrencyConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaProvisionedConcurrencyConfigQualifierAliasName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "qualifier", lambdaAliasResourceName, "name"),
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

func testAccCheckLambdaProvisionedConcurrencyConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_provisioned_concurrency_config" {
			continue
		}

		functionName, qualifier, err := resourceAwsLambdaProvisionedConcurrencyConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(input)

		if isAWSErr(err, lambda.ErrCodeProvisionedConcurrencyConfigNotFoundException, "") || isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsLambdaProvisionedConcurrencyConfigDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		functionName, qualifier, err := resourceAwsLambdaProvisionedConcurrencyConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.DeleteProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		_, err = conn.DeleteProvisionedConcurrencyConfig(input)

		return err
	}
}

func testAccCheckAwsLambdaProvisionedConcurrencyConfigExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		functionName, qualifier, err := resourceAwsLambdaProvisionedConcurrencyConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(functionName),
			Qualifier:    aws.String(qualifier),
		}

		output, err := conn.GetProvisionedConcurrencyConfig(input)

		if err != nil {
			return err
		}

		if got, want := aws.StringValue(output.Status), lambda.ProvisionedConcurrencyStatusEnumReady; got != want {
			return fmt.Errorf("Lambda Provisioned Concurrency Config (%s) expected status (%s), got: %s", rs.Primary.ID, want, got)
		}

		return nil
	}
}

func testAccAWSLambdaProvisionedConcurrencyConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
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
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.id
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdapinpoint.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambdapinpoint.handler"
  publish       = true
  runtime       = "nodejs12.x"

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccAWSLambdaProvisionedConcurrencyConfigProvisionedConcurrentExecutions(rName string, provisionedConcurrentExecutions int) string {
	return testAccAWSLambdaProvisionedConcurrencyConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = %[1]d
  qualifier                         = aws_lambda_function.test.version
}
`, provisionedConcurrentExecutions)
}

func testAccAWSLambdaProvisionedConcurrencyConfigQualifierAliasName(rName string) string {
	return testAccAWSLambdaProvisionedConcurrencyConfigBase(rName) + `
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_alias.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_alias.test.name
}
`
}

func testAccAWSLambdaProvisionedConcurrencyConfigQualifierFunctionVersion(rName string) string {
	return testAccAWSLambdaProvisionedConcurrencyConfigBase(rName) + `
resource "aws_lambda_provisioned_concurrency_config" "test" {
  function_name                     = aws_lambda_function.test.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_function.test.version
}
`
}
