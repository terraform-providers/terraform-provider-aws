package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSLambdaFunction_basic(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-basic-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-basic-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-basic-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-basic-func-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigBasic(roleName, policyName, sgName, funcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "role"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "source_code_hash"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "source_code_size"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "last_modified"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "qualified_arn"),
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "invoke_arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "function_name", funcName),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "description", funcName),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "qualifier", "$LATEST"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "handler", "exports.example"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "memory_size", "128"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "runtime", "nodejs8.10"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "timeout", "3"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "version", "$LATEST"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "reserved_concurrent_executions", "0"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "tracing_config.#", "1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "tracing_config.0.mode", "PassThrough"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_version(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-version-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-version-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-version-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-version-func-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigVersion(roleName, policyName, sgName, funcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "function_name", funcName),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "qualifier", "1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "version", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_alias(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-alias-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-alias-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-alias-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-alias-func-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigAlias(roleName, policyName, sgName, funcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "function_name", funcName),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "qualifier", "alias-name"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "version", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_layers(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-layer-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-layer-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-layer-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-layer-func-%s", rString)
	layerName := fmt.Sprintf("tf-acctest-d-lambda-function-layer-layer-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigLayers(roleName, policyName, sgName, funcName, layerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "layers.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_vpc(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-vpc-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-vpc-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-vpc-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-vpc-func-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigVPC(roleName, policyName, sgName, funcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "vpc_config.#", "1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "vpc_config.0.subnet_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLambdaFunction_environment(t *testing.T) {
	rString := acctest.RandString(7)
	roleName := fmt.Sprintf("tf-acctest-d-lambda-function-environment-role-%s", rString)
	policyName := fmt.Sprintf("tf-acctest-d-lambda-function-environment-policy-%s", rString)
	sgName := fmt.Sprintf("tf-acctest-d-lambda-function-environment-sg-%s", rString)
	funcName := fmt.Sprintf("tf-acctest-d-lambda-function-environment-func-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaFunctionConfigEnvironment(roleName, policyName, sgName, funcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_lambda_function.acctest", "arn"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "environment.#", "1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "environment.0.variables.%", "2"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "environment.0.variables.key1", "value1"),
					resource.TestCheckResourceAttr("data.aws_lambda_function.acctest", "environment.0.variables.key2", "value2"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = "%s"

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

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "lambda" {
  name = "%s"
  role = "${aws_iam_role.lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
        },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_vpc" "lambda" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "lambda" {
  vpc_id     = "${aws_vpc.lambda.id}"
  cidr_block = "10.0.1.0/24"
}

resource "aws_security_group" "lambda" {
  name        = "%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = "${aws_vpc.lambda.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, roleName, policyName, sgName)
}

func testAccDataSourceAWSLambdaFunctionConfigBasic(roleName, policyName, sgName, funcName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
}
`, funcName, funcName)
}

func testAccDataSourceAWSLambdaFunctionConfigVersion(roleName, policyName, sgName, funcName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
  publish = true
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
  qualifier     = 1
}
`, funcName, funcName)
}

func testAccDataSourceAWSLambdaFunctionConfigAlias(roleName, policyName, sgName, funcName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
  publish = true
}

resource "aws_lambda_alias" "alias" {
  name             = "alias-name"
  function_name    = "${aws_lambda_function.acctest_create.arn}"
  function_version = "1"
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
  qualifier     = "${aws_lambda_alias.alias.name}"
}
`, funcName, funcName)
}

func testAccDataSourceAWSLambdaFunctionConfigLayers(roleName, policyName, sgName, funcName, layerName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_layer_version" "acctest_create" {
  filename = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
  compatible_runtimes = ["nodejs8.10"]
}

resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
  layers = ["${aws_lambda_layer_version.acctest_create.layer_arn}"]
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
}
`, layerName, funcName, funcName)
}

func testAccDataSourceAWSLambdaFunctionConfigVPC(roleName, policyName, sgName, funcName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"

  vpc_config {
    subnet_ids = ["${aws_subnet.lambda.id}"]
    security_group_ids = ["${aws_security_group.lambda.id}"]
  }
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
}
`, funcName, funcName)
}

func testAccDataSourceAWSLambdaFunctionConfigEnvironment(roleName, policyName, sgName, funcName string) string {
	return fmt.Sprintf(testAccDataSourceAWSLambdaFunctionConfigBase(roleName, policyName, sgName)+`
resource "aws_lambda_function" "acctest_create" {
  function_name = "%s"
  description = "%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"

  environment {
    variables = {
      key1 = "value1"
      key2 = "value2"
    }
  }
}

data "aws_lambda_function" "acctest" {
  function_name = "${aws_lambda_function.acctest_create.function_name}"
}
`, funcName, funcName)
}
