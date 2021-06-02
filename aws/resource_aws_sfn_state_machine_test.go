package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sfn/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSSfnStateMachine_createUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	roleResourceName := "aws_iam_role.for_sfn"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfig(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
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
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "states", fmt.Sprintf("stateMachine:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.include_execution_data", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_expressUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineTypedConfig(rName, "EXPRESS", 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "EXPRESS"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineTypedConfig(rName, "EXPRESS", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "EXPRESS"),
				),
			},
		},
	})
}

func TestAccAWSSfnStateMachine_standardUpdate(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineTypedConfig(rName, "STANDARD", 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
				),
			},
			{
				Config: testAccAWSSfnStateMachineTypedConfig(rName, "STANDARD", 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 10.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "STANDARD"),
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
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
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

func TestAccAWSSfnStateMachine_TracingConfiguration(t *testing.T) {
	var sm sfn.DescribeStateMachineOutput
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineConfigTracingConfigDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSfnStateMachineConfigTracingConfigEnable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tracing_configuration.0.enabled", "true"),
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
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
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
	resourceName := "aws_sfn_state_machine.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sfn.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnStateMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnStateMachineExpressConfigLogConfiguration(rName, sfn.LogLevelError),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", sfn.LogLevelError),
				),
			},
			{
				Config: testAccAWSSfnStateMachineExpressConfigLogConfiguration(rName, sfn.LogLevelAll),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnExists(resourceName, &sm),
					resource.TestCheckResourceAttr(resourceName, "status", sfn.StateMachineStatusActive),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestMatchResourceAttr(resourceName, "definition", regexp.MustCompile(`.*\"MaxAttempts\": 5.*`)),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.level", sfn.LogLevelAll),
				),
			},
		},
	})
}

func testAccCheckAWSSfnExists(n string, v *sfn.DescribeStateMachineOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function State Machine ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		output, err := finder.StateMachineByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAWSSfnStateMachineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sfnconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_state_machine" {
			continue
		}

		_, err := finder.StateMachineByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Step Function State Machine %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSSfnStateMachineConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "for_lambda" {
  name = "%[1]s-lambda"
  role = aws_iam_role.for_lambda.id

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

resource "aws_iam_role" "for_lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "for_sfn" {
  name = "%[1]s-sfn"
  role = aws_iam_role.for_sfn.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
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
      "logs:DescribeLogGroups",
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets"
    ],
    "Resource": "*"
  }]
}
EOF
}

resource "aws_iam_role" "for_sfn" {
  name = "%[1]s-sfn"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "states.${data.aws_region.current.name}.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}
`, rName)
}

func testAccAWSSfnStateMachineConfig(rName string, rMaxAttempts int) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

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
          "MaxAttempts": %[2]d,
          "BackoffRate": 8
        }
      ],
      "End": true
    }
  }
}
EOF
}
`, rName, rMaxAttempts))
}

func testAccAWSSfnStateMachineConfigTags1(rName, tag1Key, tag1Value string) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

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
`, rName, tag1Key, tag1Value))
}

func testAccAWSSfnStateMachineConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

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
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccAWSSfnStateMachineTypedConfig(rName, rType string, rMaxAttempts int) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn
  type     = %[2]q

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
`, rName, rType, rMaxAttempts))
}

func testAccAWSSfnStateMachineExpressConfigLogConfiguration(rName string, rLevel string) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn
  type     = "EXPRESS"

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

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.test.arn}:*"
    include_execution_data = false
    level                  = %[2]q
  }
}
`, rName, rLevel))
}

func testAccAWSSfnStateMachineConfigTracingConfigEnable(rName string) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  tracing_configuration {
    enabled = true
  }

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
}
`, rName))
}

func testAccAWSSfnStateMachineConfigTracingConfigDisable(rName string) string {
	return composeConfig(testAccAWSSfnStateMachineConfigBase(rName), fmt.Sprintf(`
resource "aws_sfn_state_machine" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.for_sfn.arn

  tracing_configuration {
    enabled = false
  }

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
}
`, rName))
}
