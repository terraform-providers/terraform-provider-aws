package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSfnStateMachine_createUpdate(t *testing.T) {
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(name, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "definition"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfig(name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_Tags(t *testing.T) {
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigTags1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSfnExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		_, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSSfnStateMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sfnconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_state_machine" {
			continue
		}

		out, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if wserr, ok := err.(awserr.Error); ok && wserr.Code() == "StateMachineDoesNotExist" {
				return nil
			}
			return err
		}

		if out != nil && *out.Status != sfn.StateMachineStatusDeleting {
			return fmt.Errorf("Expected AWS Step Function State Machine to be destroyed, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in Step Function Test")
}

func testAccAWSSfnStateMachineConfig(rName string, rMaxAttempts int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
  current = true
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "iam_policy_for_lambda_%s"
  role = "${aws_iam_role.iam_for_lambda.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": "arn:aws:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%s"
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

resource "aws_iam_role_policy" "iam_policy_for_sfn" {
  name = "iam_policy_for_sfn_%s"
  role = "${aws_iam_role.iam_for_sfn.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
        "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
  filename = "test-fixtures/lambdatest.zip"
  function_name = "sfn-%s"
  role = "${aws_iam_role.iam_for_lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%s"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda_function_test.arn}",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": %d,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}

`, rName, rName, rName, rName, rName, rName, rMaxAttempts)
}

func testAccAWSSfnStateMachineConfigTags1(rName string, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
  current = true
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "iam_policy_for_lambda_%s"
  role = "${aws_iam_role.iam_for_lambda.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": "arn:aws:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%s"
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

resource "aws_iam_role_policy" "iam_policy_for_sfn" {
  name = "iam_policy_for_sfn_%s"
  role = "${aws_iam_role.iam_for_sfn.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
        "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
  filename = "test-fixtures/lambdatest.zip"
  function_name = "sfn-%s"
  role = "${aws_iam_role.iam_for_lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%s"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda_function_test.arn}",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
tags = {
	%q = %q
}
}
`, rName, rName, rName, rName, rName, rName, tag1Key, tag1Value)
}

func testAccAWSSfnStateMachineConfigTags2(rName string, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
  current = true
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "iam_policy_for_lambda_%s"
  role = "${aws_iam_role.iam_for_lambda.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": "arn:aws:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%s"
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

resource "aws_iam_role_policy" "iam_policy_for_sfn" {
  name = "iam_policy_for_sfn_%s"
  role = "${aws_iam_role.iam_for_sfn.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
        "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
  filename = "test-fixtures/lambdatest.zip"
  function_name = "sfn-%s"
  role = "${aws_iam_role.iam_for_lambda.arn}"
  handler = "exports.example"
  runtime = "nodejs8.10"
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%s"
  role_arn = "${aws_iam_role.iam_for_sfn.arn}"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda_function_test.arn}",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
tags = {
	%q = %q
	%q = %q
}
}
`, rName, rName, rName, rName, rName, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
