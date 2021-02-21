package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSfnStateMachine_createUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	roleResourceName := "aws_iam_role.iam_for_sfn"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "states", regexp.MustCompile(`stateMachine:.+`)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "states", regexp.MustCompile(`stateMachine:.+`)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_expressUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineTypedConfig(sfn.StateMachineTypeExpress, name, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "definition"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "type", "EXPRESS"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineTypedConfig(sfn.StateMachineTypeExpress, name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "type", "EXPRESS"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_standardUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineTypedConfig(sfn.StateMachineTypeStandard, name, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "definition"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "type", "STANDARD"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineTypedConfig(sfn.StateMachineTypeStandard, name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "type", "STANDARD"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_tags(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_disappears(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSfnStateMachine(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSfnStateMachine_expressLoggingConfiguration(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineExpressConfigLogConfiguration(sfn.StateMachineTypeExpress, name, sfn.LogLevelError),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "definition"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "logging_configuration.0.level", sfn.LogLevelError),
				),
			},
			{
				Config: testAccAWSSfnStateMachineExpressConfigLogConfiguration(sfn.StateMachineTypeExpress, name, sfn.LogLevelAll),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists("aws_sfn_state_machine.foo", &sm),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "name"),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "creation_date"),
					resource.TestMatchResourceAttr("aws_sfn_state_machine.foo", "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet("aws_sfn_state_machine.foo", "role_arn"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_sfn_state_machine.foo", "logging_configuration.0.level", sfn.LogLevelAll),
				),
			},
		},
	})
}

func testAccCheckAWSSfnExists(n string, sm *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		resp, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*sm = *resp

		return nil
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
			if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "") {
				continue
			}
		}

		if out != nil && aws.StringValue(out.Status) != sfn.StateMachineStatusDeleting {
			return fmt.Errorf("Expected AWS Step Function State Machine to be destroyed, but was still found")
		}

		return err
	}

	return nil
}

func testAccAWSSfnStateMachineConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%[1]s-2"

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
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

data "aws_region" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_sfn" {
  name = %[1]q
  role = aws_iam_role.iam_for_sfn.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": "${aws_lambda_function.test.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "states.${data.aws_region.current.name}.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSSfnStateMachineConfig(rName string, rMaxAttempts int) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %q
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": %d,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rMaxAttempts)
}

func testAccAWSSfnStateMachineConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSSfnStateMachineConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAWSSfnStateMachineConfigBase(rName) + fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.test.arn}",
      "Retry": [
        {
          "ErrorEquals": [
            "States.ALL"
          ],
          "IntervalSeconds": 5,
          "MaxAttempts": 5,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSSfnStateMachineTypedConfig(rType string, rName string, rMaxAttempts int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "iam_policy_for_lambda_%[1]s"
  role = aws_iam_role.iam_for_lambda.id

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
    "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%[1]s"

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
  name = "iam_policy_for_sfn_%[1]s"
  role = aws_iam_role.iam_for_sfn.id

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
  name = "iam_for_sfn_%[1]s"

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
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "sfn-%[1]s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%[1]s"
  role_arn = aws_iam_role.iam_for_sfn.arn
  type     = %[2]q

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function with %[2]s State Machine",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda_function_test.arn}",
      "Retry": [
        {
          "ErrorEquals": ["States.ALL"],
          "IntervalSeconds": 5,
          "MaxAttempts": %[3]d,
          "BackoffRate": 8.0
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rType, rMaxAttempts)
}

func testAccAWSSfnStateMachineExpressConfigLogConfiguration(rType string, rName string, rLevel string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "iam_policy_for_lambda_%[1]s"
  role = aws_iam_role.iam_for_lambda.id

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
    "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
  }]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda_%[1]s"

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
  name = "iam_policy_for_sfn_%[1]s"
  role = aws_iam_role.iam_for_sfn.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction",
        "logs:CreateLogDelivery",
        "logs:GetLogDelivery",
        "logs:UpdateLogDelivery",
        "logs:DeleteLogDelivery",
        "logs:ListLogDeliveries",
        "logs:PutResourcePolicy",
        "logs:DescribeResourcePolicies",
        "logs:DescribeLogGroups"
      ],
        "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_sfn" {
  name = "iam_for_sfn_%[1]s"

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
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "sfn-%[1]s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_cloudwatch_log_group" "log_group_for_sfn" {
  name = "log_group_sfn_%[1]s"
}

resource "aws_sfn_state_machine" "foo" {
  name     = "test_sfn_%[1]s"
  role_arn = aws_iam_role.iam_for_sfn.arn
  type     = %[2]q

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function with %[2]s State Machine",
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

  logging_configuration {
    log_destination        = aws_cloudwatch_log_group.log_group_for_sfn.arn
    include_execution_data = false
    level                  = %[3]q
  }

  tags = {
    "key1" = "value1"
  }
}
`, rName, rType, rLevel)
}
