package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCustomResource_basic(t *testing.T) {
	resourceName := "aws_custom_resource.custom_resource"
	rType := "CustomResource"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCustomResourceDestroy, //theres nothing to cleanup
		Steps: []resource.TestStep{
			{
				Config: testAccCustomResourceConfig(rType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomResource(resourceName),
				),
			},
		},
	})
}

func testAccCheckAwsCustomResourceDestroy(s *terraform.State) error {
	//there is nothing to delete because Custom Resource doesn't actually provision anything
	return nil

}

func testAccCheckCustomResource(customResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[customResource]
		if !ok {
			return fmt.Errorf("Not found: %s", customResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccCustomResourceConfig(resourceType string) string {
	return fmt.Sprintf(`
	resource "aws_iam_role" "iam_for_lambda" {
		name = "iam_for_lambda"
		force_detach_policies = true
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
	  
	  resource "aws_iam_policy" "lambda_logging" {
		name = "lambda_logging"
		path = "/"
		description = "IAM policy for logging from a lambda"
	  
		policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Action": [
		"logs:CreateLogStream",
		"logs:PutLogEvents"
	],
	"Resource": "arn:aws:logs:*:*:*",
	"Effect": "Allow"
	}
]
}
EOF
	  }
	  
	  resource "aws_iam_role_policy_attachment" "lambda_logs" {
		role = aws_iam_role.iam_for_lambda.name
		policy_arn = aws_iam_policy.lambda_logging.arn
	  }
	  
	  resource "aws_lambda_function" "lambda_function" {
		filename      = "test-fixtures/custom_resource.zip"
		function_name = "custom_resource_lambda"
		role          = aws_iam_role.iam_for_lambda.arn
		handler       = "index.handler"
		runtime       = "nodejs10.x"
	  }
	  
	  resource "aws_custom_resource" "custom_resource" {
		service_token = aws_lambda_function.lambda_function.arn
		resource_type = "%s"
		resource_properties = {
		  a = "1"
		  b = "2"
		}
	  }
`, resourceType)
}
