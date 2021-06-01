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

func TestAccAWSLambdaFunctionEventInvokeConfig_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
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

func TestAccAWSLambdaFunctionEventInvokeConfig_disappears_LambdaFunction(t *testing.T) {
	var function lambda.GetFunctionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionExists(lambdaFunctionResourceName, rName, &function),
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					testAccCheckAwsLambdaFunctionDisappears(&function),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_disappears_LambdaFunctionEventInvokeConfig(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					testAccCheckAwsLambdaFunctionEventInvokeConfigDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_DestinationConfig_OnFailure_Destination(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSqsQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSnsTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", snsTopicResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_DestinationConfig_OnSuccess_Destination(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnSuccessDestinationSqsQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnSuccessDestinationSnsTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", snsTopicResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_DestinationConfig_Remove(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSqsQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigQualifierFunctionVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_DestinationConfig_Swap(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"
	sqsQueueResourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSqsQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination", sqsQueueResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnSuccessDestinationSqsQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_success.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_success.0.destination", sqsQueueResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_FunctionName_Arn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigFunctionNameArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
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

func TestAccAWSLambdaFunctionEventInvokeConfig_Qualifier_FunctionName_Arn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigQualifierFunctionNameArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", LambdaFunctionVersionLatest),
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

func TestAccAWSLambdaFunctionEventInvokeConfig_MaximumEventAgeInSeconds(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigMaximumEventAgeInSeconds(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigMaximumEventAgeInSeconds(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_event_age_in_seconds", "200"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_MaximumRetryAttempts(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigMaximumRetryAttempts(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigMaximumRetryAttempts(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "1"),
				),
			},
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigMaximumRetryAttempts(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", "0"),
				),
			},
		},
	})
}

func TestAccAWSLambdaFunctionEventInvokeConfig_Qualifier_AliasName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigQualifierAliasName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
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

func TestAccAWSLambdaFunctionEventInvokeConfig_Qualifier_FunctionVersion(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigQualifierFunctionVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", lambdaFunctionResourceName, "function_name"),
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

func TestAccAWSLambdaFunctionEventInvokeConfig_Qualifier_Latest(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_function_event_invoke_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaFunctionEventInvokeConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaFunctionEventInvokeConfigQualifierLatest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "qualifier", LambdaFunctionVersionLatest),
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

func testAccCheckLambdaFunctionEventInvokeConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function_event_invoke_config" {
			continue
		}

		functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		output, err := conn.GetFunctionEventInvokeConfig(input)

		if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("Lambda Function Event Invoke Config (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsLambdaFunctionEventInvokeConfigDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.DeleteFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		_, err = conn.DeleteFunctionEventInvokeConfig(input)

		return err
	}
}

func testAccCheckAwsLambdaFunctionEventInvokeConfigExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &lambda.GetFunctionEventInvokeConfigInput{
			FunctionName: aws.String(functionName),
		}

		if qualifier != "" {
			input.Qualifier = aws.String(qualifier)
		}

		_, err = conn.GetFunctionEventInvokeConfig(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSLambdaFunctionEventInvokeConfigBase(rName string) string {
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

  depends_on = [
    aws_iam_role_policy_attachment.test,
  ]
}
`, rName)
}

func testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSnsTopic(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSNSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSNSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_failure {
      destination = aws_sns_topic.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSNSFullAccess]
}
`, rName)
}

func testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnFailureDestinationSqsQueue(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSQSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSQSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSQSFullAccess]
}
`, rName)
}

func testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnSuccessDestinationSnsTopic(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSNSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSNSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_success {
      destination = aws_sns_topic.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSNSFullAccess]
}
`, rName)
}

func testAccAWSLambdaFunctionEventInvokeConfigDestinationConfigOnSuccessDestinationSqsQueue(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonSQSFullAccess" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSQSFullAccess"
  role       = aws_iam_role.test.id
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name

  destination_config {
    on_success {
      destination = aws_sqs_queue.test.arn
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonSQSFullAccess]
}
`, rName)
}

func testAccAWSLambdaFunctionEventInvokeConfigFunctionName(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
}
`
}

func testAccAWSLambdaFunctionEventInvokeConfigFunctionNameArn(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
}
`
}

func testAccAWSLambdaFunctionEventInvokeConfigQualifierFunctionNameArn(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.arn
  qualifier     = "$LATEST"
}
`
}

func testAccAWSLambdaFunctionEventInvokeConfigMaximumEventAgeInSeconds(rName string, maximumEventAgeInSeconds int) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name                = aws_lambda_function.test.function_name
  maximum_event_age_in_seconds = %[1]d
}
`, maximumEventAgeInSeconds)
}

func testAccAWSLambdaFunctionEventInvokeConfigMaximumRetryAttempts(rName string, maximumRetryAttempts int) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + fmt.Sprintf(`
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name          = aws_lambda_function.test.function_name
  maximum_retry_attempts = %[1]d
}
`, maximumRetryAttempts)
}

func testAccAWSLambdaFunctionEventInvokeConfigQualifierAliasName(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_alias" "test" {
  function_name    = aws_lambda_function.test.function_name
  function_version = aws_lambda_function.test.version
  name             = "test"
}

resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_alias.test.function_name
  qualifier     = aws_lambda_alias.test.name
}
`
}

func testAccAWSLambdaFunctionEventInvokeConfigQualifierFunctionVersion(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_function.test.version
}
`
}

func testAccAWSLambdaFunctionEventInvokeConfigQualifierLatest(rName string) string {
	return testAccAWSLambdaFunctionEventInvokeConfigBase(rName) + `
resource "aws_lambda_function_event_invoke_config" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = "$LATEST"
}
`
}
