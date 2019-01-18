package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudwatchLogSubscriptionFilter_basic(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter

	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudwatchLogSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchLogSubscriptionFilterConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudwatchLogSubscriptionFilterExists("aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", &filter, rstring),
					resource.TestCheckResourceAttr(
						"aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", "filter_pattern", "logtype test"),
					resource.TestCheckResourceAttr(
						"aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", "name", fmt.Sprintf("test_lambdafunction_logfilter_%s", rstring)),
					resource.TestCheckResourceAttr(
						"aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", "log_group_name", fmt.Sprintf("example_lambda_name_%s", rstring)),
					resource.TestCheckResourceAttr(
						"aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", "distribution", "Random"),
				),
			},
		},
	})
}

func TestAccAWSCloudwatchLogSubscriptionFilter_disappears(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudwatchLogSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchLogSubscriptionFilterConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudwatchLogSubscriptionFilterExists("aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter", &filter, rstring),
					testAccCheckCloudwatchLogSubscriptionFilterDisappears(rstring),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudwatchLogSubscriptionFilter_disappears_LogGroup(t *testing.T) {
	var filter cloudwatchlogs.SubscriptionFilter
	var logGroup cloudwatchlogs.LogGroup

	rstring := acctest.RandString(5)
	logGroupResourceName := "aws_cloudwatch_log_group.logs"
	resourceName := "aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudwatchLogSubscriptionFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchLogSubscriptionFilterConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudwatchLogSubscriptionFilterExists(resourceName, &filter, rstring),
					testAccCheckCloudWatchLogGroupExists(logGroupResourceName, &logGroup),
					testAccCheckCloudWatchLogGroupDisappears(&logGroup),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudwatchLogSubscriptionFilterDisappears(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		params := &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: aws.String(fmt.Sprintf("example_lambda_name_%s", rName)),
		}
		_, err := conn.DeleteLogGroup(params)
		return err
	}
}

func testAccCheckCloudwatchLogSubscriptionFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_subscription_filter" {
			continue
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		filterNamePrefix := rs.Primary.Attributes["name"]

		input := cloudwatchlogs.DescribeSubscriptionFiltersInput{
			LogGroupName:     aws.String(logGroupName),
			FilterNamePrefix: aws.String(filterNamePrefix),
		}

		_, err := conn.DescribeSubscriptionFilters(&input)
		if err == nil {
			return fmt.Errorf("SubscriptionFilter still exists")
		}

	}

	return nil

}

func testAccCheckAwsCloudwatchLogSubscriptionFilterExists(n string, filter *cloudwatchlogs.SubscriptionFilter, rName string) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SubscriptionFilter not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SubscriptionFilter ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

		logGroupName := rs.Primary.Attributes["log_group_name"]
		filterNamePrefix := rs.Primary.Attributes["name"]

		input := cloudwatchlogs.DescribeSubscriptionFiltersInput{
			LogGroupName:     aws.String(logGroupName),
			FilterNamePrefix: aws.String(filterNamePrefix),
		}

		resp, err := conn.DescribeSubscriptionFilters(&input)
		if err == nil {
			return err
		}

		if len(resp.SubscriptionFilters) == 0 {
			return fmt.Errorf("SubscriptionFilter not found")
		}

		var found bool
		for _, i := range resp.SubscriptionFilters {
			if *i.FilterName == fmt.Sprintf("test_lambdafunction_logfilter_%s", rName) {
				*filter = *i
				found = true
			}
		}

		if !found {
			return fmt.Errorf("SubscriptionFilter not found")
		}

		return nil
	}
}

func testAccAWSCloudwatchLogSubscriptionFilterConfig(rstring string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_subscription_filter" "test_lambdafunction_logfilter" {
  name            = "test_lambdafunction_logfilter_%s"
  log_group_name  = "${aws_cloudwatch_log_group.logs.name}"
  filter_pattern  = "logtype test"
  destination_arn = "${aws_lambda_function.test_lambdafunction.arn}"
  distribution    = "Random"
}

resource "aws_lambda_function" "test_lambdafunction" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "example_lambda_name_%s"
  role          = "${aws_iam_role.iam_for_lambda.arn}"
  runtime       = "nodejs8.10"
  handler       = "exports.handler"
}

resource "aws_cloudwatch_log_group" "logs" {
  name              = "example_lambda_name_%s"
  retention_in_days = 1
}

resource "aws_lambda_permission" "allow_cloudwatch_logs" {
  statement_id  = "AllowExecutionFromCloudWatchLogs"
  action        = "lambda:*"
  function_name = "${aws_lambda_function.test_lambdafunction.arn}"
  principal     = "logs.amazonaws.com"
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "test_lambdafuntion_iam_role_%s"

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

resource "aws_iam_role_policy" "test_lambdafunction_iam_policy" {
  name = "test_lambdafunction_iam_policy_%s"
  role = "${aws_iam_role.iam_for_lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1441111030000",
      "Effect": "Allow",
      "Action": [
        "lambda:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rstring, rstring, rstring, rstring, rstring)
}
