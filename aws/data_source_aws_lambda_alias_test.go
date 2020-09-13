package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSLambdaAlias_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	dataSourceName := "data.aws_lambda_alias.test"
	resourceName := "aws_lambda_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLambdaAliasConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_version", resourceName, "function_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "invoke_arn", resourceName, "invoke_arn"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLambdaAliasConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "lambda" {
  name = %[1]q
  role = aws_iam_role.lambda.id

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

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  publish       = true
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_alias" "test" {
  name             = "test"
  function_name    = aws_lambda_function.test.function_name
  function_version = "1"
}
`, rName)
}

func testAccDataSourceAWSLambdaAliasConfigBasic(rName string) string {
	return testAccDataSourceAWSLambdaAliasConfigBase(rName) + `
data "aws_lambda_alias" "test" {
  name          = aws_lambda_alias.test.name
  function_name = aws_lambda_alias.test.function_name
}
`
}
