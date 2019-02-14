package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestSuppressEquivalentSnsTopicSubscriptionDeliveryPolicy(t *testing.T) {
	var testCases = []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"maxDelayTarget":20,"minDelayTarget":5,"numRetries":5}}`,
			equivalent: true,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5}}`,
			equivalent: true,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":6}}`,
			equivalent: false,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20}}`,
			equivalent: false,
		},
		{
			old:        `{"healthyRetryPolicy":null,"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":true}`,
			new:        `{"guaranteed":true}`,
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		actual := suppressEquivalentSnsTopicSubscriptionDeliveryPolicy("", tc.old, tc.new, nil)
		if actual != tc.equivalent {
			t.Fatalf("Test Case %d: Got: %t Expected: %t", i, actual, tc.equivalent)
		}
	}
}

func TestAccAWSSNSTopicSubscription_basic(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sns", regexp.MustCompile(fmt.Sprintf("terraform-test-topic-%d:.+", ri))),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", "aws_sqs_queue.test_queue", "arn"),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "protocol", "sqs"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "topic_arn", "aws_sns_topic.test_topic", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_filterPolicy(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()
	filterPolicy1 := `{"key1": ["val1"], "key2": ["val2"]}`
	filterPolicy2 := `{"key3": ["val3"], "key4": ["val4"]}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_filterPolicy(ri, strconv.Quote(filterPolicy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", filterPolicy1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_filterPolicy(ri, strconv.Quote(filterPolicy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", filterPolicy2),
				),
			},
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_deliveryPolicy(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(ri, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries": 5}}`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckAWSSNSTopicSubscriptionDeliveryPolicyAttribute(attributes, &snsTopicSubscriptionDeliveryPolicy{
						HealthyRetryPolicy: &snsTopicSubscriptionDeliveryPolicyHealthyRetryPolicy{
							MaxDelayTarget: 20,
							MinDelayTarget: 5,
							NumRetries:     5,
						},
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(ri, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":3,"maxDelayTarget":78,"numRetries": 11}}`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckAWSSNSTopicSubscriptionDeliveryPolicyAttribute(attributes, &snsTopicSubscriptionDeliveryPolicy{
						HealthyRetryPolicy: &snsTopicSubscriptionDeliveryPolicyHealthyRetryPolicy{
							MaxDelayTarget: 78,
							MinDelayTarget: 3,
							NumRetries:     11,
						},
					}),
				),
			},
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_rawMessageDelivery(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
				),
			},
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_autoConfirmingEndpoint(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_autoConfirmingEndpoint(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_autoConfirmingSecuredEndpoint(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test_subscription"
	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(ri, "john", "doe"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func testAccCheckAWSSNSTopicSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).snsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic" {
			continue
		}

		// Try to find key pair
		req := &sns.GetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetSubscriptionAttributes(req)

		if err == nil {
			return fmt.Errorf("Subscription still exists, can't continue.")
		}

		// Verify the error is an API error, not something else
		_, ok := err.(awserr.Error)
		if !ok {
			return err
		}
	}

	return nil
}

func testAccCheckAWSSNSTopicSubscriptionExists(n string, attributes map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS subscription with that ARN exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).snsconn

		params := &sns.GetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetSubscriptionAttributes(params)

		for k, v := range output.Attributes {
			attributes[k] = aws.StringValue(v)
		}

		return err
	}
}

func testAccCheckAWSSNSTopicSubscriptionDeliveryPolicyAttribute(attributes map[string]string, expectedDeliveryPolicy *snsTopicSubscriptionDeliveryPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiDeliveryPolicyJSONString, ok := attributes["DeliveryPolicy"]

		if !ok {
			return fmt.Errorf("DeliveryPolicy attribute not found in attributes: %s", attributes)
		}

		var apiDeliveryPolicy snsTopicSubscriptionDeliveryPolicy
		if err := json.Unmarshal([]byte(apiDeliveryPolicyJSONString), &apiDeliveryPolicy); err != nil {
			return fmt.Errorf("unable to unmarshal SNS Topic Subscription delivery policy JSON (%s): %s", apiDeliveryPolicyJSONString, err)
		}

		if reflect.DeepEqual(apiDeliveryPolicy, *expectedDeliveryPolicy) {
			return nil
		}

		return fmt.Errorf("SNS Topic Subscription delivery policy did not match:\n\nReceived\n\n%s\n\nExpected\n\n%s\n\n", apiDeliveryPolicy, *expectedDeliveryPolicy)
	}
}

func TestObfuscateEndpointPassword(t *testing.T) {
	checks := map[string]string{
		"https://example.com/myroute":                   "https://example.com/myroute",
		"https://username@example.com/myroute":          "https://username@example.com/myroute",
		"https://username:password@example.com/myroute": "https://username:****@example.com/myroute",
	}
	for endpoint, expected := range checks {
		out := obfuscateEndpoint(endpoint)

		if expected != out {
			t.Fatalf("Expected %v, got %v", expected, out)
		}
	}
}

func testAccAWSSNSTopicSubscriptionConfig(i int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "terraform-test-topic-%d"
}

resource "aws_sqs_queue" "test_queue" {
  name = "terraform-subscription-test-queue-%d"
}

resource "aws_sns_topic_subscription" "test_subscription" {
  topic_arn = "${aws_sns_topic.test_topic.arn}"
  protocol = "sqs"
  endpoint = "${aws_sqs_queue.test_queue.arn}"
}
`, i, i)
}

func testAccAWSSNSTopicSubscriptionConfig_filterPolicy(i int, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "terraform-test-topic-%d"
}

resource "aws_sqs_queue" "test_queue" {
  name = "terraform-subscription-test-queue-%d"
}

resource "aws_sns_topic_subscription" "test_subscription" {
  topic_arn = "${aws_sns_topic.test_topic.arn}"
  protocol = "sqs"
  endpoint = "${aws_sqs_queue.test_queue.arn}"
  filter_policy = %s
}
`, i, i, policy)
}

func testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(i int, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "terraform-test-topic-%d"
}

resource "aws_sqs_queue" "test_queue" {
  name = "terraform-subscription-test-queue-%d"
}

resource "aws_sns_topic_subscription" "test_subscription" {
  delivery_policy = %s
  endpoint        = "${aws_sqs_queue.test_queue.arn}"
  protocol        = "sqs"
  topic_arn       = "${aws_sns_topic.test_topic.arn}"
}
`, i, i, policy)
}

func testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(i int, rawMessageDelivery bool) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "terraform-test-topic-%d"
}

resource "aws_sqs_queue" "test_queue" {
  name = "terraform-subscription-test-queue-%d"
}

resource "aws_sns_topic_subscription" "test_subscription" {
  endpoint             = "${aws_sqs_queue.test_queue.arn}"
  protocol             = "sqs"
  raw_message_delivery = %t
  topic_arn            = "${aws_sns_topic.test_topic.arn}"
}
`, i, i, rawMessageDelivery)
}

func testAccAWSSNSTopicSubscriptionConfig_autoConfirmingEndpoint(i int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "tf-acc-test-sns-%d"
}

resource "aws_api_gateway_rest_api" "test" {
  name        = "tf-acc-test-sns-%d"
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  resource_id   = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = "${aws_api_gateway_rest_api.test.id}"
  resource_id             = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method             = "${aws_api_gateway_method.test.http_method}"
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "${aws_lambda_function.lambda.invoke_arn}"
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = ["aws_api_gateway_integration.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  status_code = "${aws_api_gateway_method_response.test.status_code}"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "tf-acc-test-sns-%d"

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

resource "aws_iam_role_policy" "policy" {
  name = "tf-acc-test-sns-%d"
  role = "${aws_iam_role.iam_for_lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.lambda.arn}"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = "tf-acc-test-sns-%d"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "main.confirm_subscription"
  source_code_hash = "${base64sha256(file("test-fixtures/lambda_confirm_sns.zip"))}"
  runtime          = "python3.6"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = ["aws_api_gateway_integration_response.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "acctest"
}

resource "aws_sns_topic_subscription" "test_subscription" {
  depends_on             = ["aws_lambda_permission.apigw_lambda"]
  topic_arn              = "${aws_sns_topic.test_topic.arn}"
  protocol               = "https"
  endpoint               = "${aws_api_gateway_deployment.test.invoke_url}"
  endpoint_auto_confirms = true
}
`, i, i, i, i, i)
}

func testAccAWSSNSTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(i int, username, password string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "tf-acc-test-sns-%d"
}

resource "aws_api_gateway_rest_api" "test" {
  name        = "tf-acc-test-sns-%d"
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = "${aws_api_gateway_rest_api.test.id}"
  resource_id   = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method   = "POST"
  authorization = "CUSTOM"
  authorizer_id = "${aws_api_gateway_authorizer.test.id}"
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = "${aws_api_gateway_rest_api.test.id}"
  resource_id             = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method             = "${aws_api_gateway_method.test.http_method}"
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "${aws_lambda_function.lambda.invoke_arn}"
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = ["aws_api_gateway_integration.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  resource_id = "${aws_api_gateway_rest_api.test.root_resource_id}"
  http_method = "${aws_api_gateway_method.test.http_method}"
  status_code = "${aws_api_gateway_method_response.test.status_code}"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "tf-acc-test-sns-%d"

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

resource "aws_iam_role_policy" "policy" {
  name = "tf-acc-test-sns-%d"
  role = "${aws_iam_role.iam_for_lambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.lambda.arn}"
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = "tf-acc-test-sns-%d"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "main.confirm_subscription"
  source_code_hash = "${base64sha256(file("test-fixtures/lambda_confirm_sns.zip"))}"
  runtime          = "python3.6"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = ["aws_api_gateway_integration_response.test"]
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  stage_name  = "acctest"
}

resource "aws_iam_role" "invocation_role" {
  name = "tf-acc-test-authorizer-%d"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "tf-acc-test-authorizer-%d"
  role = "${aws_iam_role.invocation_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer.arn}"
    }
  ]
}
EOF
}

resource "aws_api_gateway_authorizer" "test" {
  name                             = "tf-acc-test-api-gw-authorizer-%d"
  rest_api_id                      = "${aws_api_gateway_rest_api.test.id}"
  authorizer_uri                   = "${aws_lambda_function.authorizer.invoke_arn}"
  authorizer_result_ttl_in_seconds = "0"
  authorizer_credentials           = "${aws_iam_role.invocation_role.arn}"
}

resource "aws_lambda_function" "authorizer" {
  filename         = "test-fixtures/lambda_basic_authorizer.zip"
  source_code_hash = "${base64sha256(file("test-fixtures/lambda_basic_authorizer.zip"))}"
  function_name    = "tf-acc-test-authorizer-%d"
  role             = "${aws_iam_role.iam_for_lambda.arn}"
  handler          = "main.authenticate"
  runtime          = "nodejs6.10"

  environment {
    variables = {
        AUTH_USER="%s"
        AUTH_PASS="%s"
    }
  }
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id = "${aws_api_gateway_rest_api.test.id}"
  status_code = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.WWW-Authenticate" = "'Basic'"
  }
}

resource "aws_sns_topic_subscription" "test_subscription" {
  depends_on             = ["aws_lambda_permission.apigw_lambda"]
  topic_arn              = "${aws_sns_topic.test_topic.arn}"
  protocol               = "https"
  endpoint               = "${replace(aws_api_gateway_deployment.test.invoke_url, "https://", "https://%s:%s@")}"
  endpoint_auto_confirms = true
}
`, i, i, i, i, i, i, i, i, i, username, password, username, password)
}
