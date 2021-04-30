package aws

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sns/finder"
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
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"throttlePolicy":{}}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"throttlePolicy":{}}`,
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
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", sns.ServiceName, regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "confirmation_was_authenticated", "true"),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", "aws_sqs_queue.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_confirmation", "false"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "sqs"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "topic_arn", "aws_sns_topic.test", "arn"),
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
	resourceName := "aws_sns_topic_subscription.test"
	filterPolicy1 := `{"key1": ["val1"], "key2": ["val2"]}`
	filterPolicy2 := `{"key3": ["val3"], "key4": ["val4"]}`
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_filterPolicy(rName, strconv.Quote(filterPolicy1)),
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
				Config: testAccAWSSNSTopicSubscriptionConfig_filterPolicy(rName, strconv.Quote(filterPolicy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", filterPolicy2),
				),
			},
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
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
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(rName, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries": 5}}`)),
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
				Config: testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(rName, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":3,"maxDelayTarget":78,"numRetries": 11}}`)),
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
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_redrivePolicy(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	dlqName := acctest.RandomWithPrefix("tf-acc-test")
	updatedDlqName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_redrivePolicy(rName, dlqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckAWSSNSTopicSubscriptionRedrivePolicyAttribute(attributes, dlqName),
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
				Config: testAccAWSSNSTopicSubscriptionConfig_redrivePolicy(rName, updatedDlqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckAWSSNSTopicSubscriptionRedrivePolicyAttribute(attributes, updatedDlqName),
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
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_rawMessageDelivery(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(rName, true),
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
				Config: testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
				),
			},
			// Test attribute removal
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
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
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_autoConfirmingEndpoint(rName),
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
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(rName),
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

func TestAccAWSSNSTopicSubscription_email(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionEmailConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", sns.ServiceName, regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "confirmation_was_authenticated", "false"),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "endpoint", "invalid_email@example.com"),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_confirmation", "true"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "email"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "topic_arn", "aws_sns_topic.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_firehose(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", sns.ServiceName, regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", "aws_kinesis_firehose_delivery_stream.test_stream", "arn"),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "protocol", "firehose"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_role_arn", "aws_iam_role.firehose_role", "arn"),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_disappears(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSnsTopicSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSNSTopicSubscription_disappears_topic(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sns.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicSubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicSubscriptionExists(resourceName, attributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSnsTopic(), "aws_sns_topic.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSNSTopicSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).snsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic_subscription" {
			continue
		}

		output, err := finder.SubscriptionByARN(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("SNS topic subscription still exists, can't continue.")
		}

		if output == nil || aws.StringValue(output.Attributes["Protocol"]) == "email" {
			return nil
		}

		return fmt.Errorf("SNS topic Subscription (%s) exists when it should be destroyed", rs.Primary.ID)
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

		output, err := finder.SubscriptionByARN(conn, rs.Primary.ID)
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

func testAccCheckAWSSNSTopicSubscriptionRedrivePolicyAttribute(attributes map[string]string, expectedRedrivePolicyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiRedrivePolicyJSONString, ok := attributes["RedrivePolicy"]

		if !ok {
			return fmt.Errorf("RedrivePolicy attribute not found in attributes: %s", attributes)
		}

		var apiRedrivePolicy snsTopicSubscriptionRedrivePolicy
		if err := json.Unmarshal([]byte(apiRedrivePolicyJSONString), &apiRedrivePolicy); err != nil {
			return fmt.Errorf("unable to unmarshal SNS Topic Subscription redrive policy JSON (%s): %s", apiRedrivePolicyJSONString, err)
		}

		expectedRedrivePolicy := snsTopicSubscriptionRedrivePolicy{
			DeadLetterTargetArn: arn.ARN{
				AccountID: testAccGetAccountID(),
				Partition: testAccGetPartition(),
				Region:    testAccGetRegion(),
				Resource:  expectedRedrivePolicyResource,
				Service:   "sqs",
			}.String(),
		}

		if reflect.DeepEqual(apiRedrivePolicy, expectedRedrivePolicy) {
			return nil
		}

		return fmt.Errorf("SNS Topic Subscription redrive policy did not match:\n\nReceived\n\n%s\n\nExpected\n\n%s\n\n", apiRedrivePolicy, expectedRedrivePolicy)
	}
}

const awsSNSPasswordObfuscationPattern = "****"

// returns the endpoint with obfuscated password, if any
func obfuscateEndpoint(t *testing.T, endpoint string) string {
	res, err := url.Parse(endpoint)
	if err != nil {
		t.Errorf("error parsing URL: %s", err)
	}

	var obfuscatedEndpoint = res.String()

	// If the user is defined, we try to get the username and password, if defined.
	// Then, we update the user with the obfuscated version.
	if res.User != nil {
		if password, ok := res.User.Password(); ok {
			obfuscatedEndpoint = strings.Replace(obfuscatedEndpoint, password, awsSNSPasswordObfuscationPattern, 1)
		}
	}
	return obfuscatedEndpoint
}

func TestObfuscateEndpointPassword(t *testing.T) {
	checks := map[string]string{
		"https://example.com/myroute":                   "https://example.com/myroute",
		"https://username@example.com/myroute":          "https://username@example.com/myroute",
		"https://username:password@example.com/myroute": "https://username:****@example.com/myroute",
	}
	for endpoint, expected := range checks {
		out := obfuscateEndpoint(t, endpoint)

		if expected != out {
			t.Fatalf("Expected %v, got %v", expected, out)
		}
	}
}

func testAccAWSSNSTopicSubscriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccAWSSNSTopicSubscriptionConfig_filterPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn     = aws_sns_topic.test.arn
  protocol      = "sqs"
  endpoint      = aws_sqs_queue.test.arn
  filter_policy = %[2]s
}
`, rName, policy)
}

func testAccAWSSNSTopicSubscriptionConfig_deliveryPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  delivery_policy = %[2]s
  endpoint        = aws_sqs_queue.test.arn
  protocol        = "sqs"
  topic_arn       = aws_sns_topic.test.arn
}
`, rName, policy)
}

func testAccAWSSNSTopicSubscriptionConfig_redrivePolicy(rName, dlqName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test_dlq" {
  name = %[2]q
}

resource "aws_sns_topic_subscription" "test" {
  redrive_policy = jsonencode({ deadLetterTargetArn : aws_sqs_queue.test_dlq.arn })
  endpoint       = aws_sqs_queue.test.arn
  protocol       = "sqs"
  topic_arn      = aws_sns_topic.test.arn
}
`, rName, dlqName)
}

func testAccAWSSNSTopicSubscriptionConfig_rawMessageDelivery(rName string, rawMessageDelivery bool) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  endpoint             = aws_sqs_queue.test.arn
  protocol             = "sqs"
  raw_message_delivery = %[2]t
  topic_arn            = aws_sns_topic.test.arn
}
`, rName, rawMessageDelivery)
}

func testAccAWSSNSTopicSubscriptionConfig_autoConfirmingEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api" "test" {
  name        = %[1]q
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_rest_api.test.root_resource_id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = aws_api_gateway_rest_api.test.id
  resource_id             = aws_api_gateway_rest_api.test.root_resource_id
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = [aws_api_gateway_integration.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "iam_for_lambda" {
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

resource "aws_iam_role_policy" "policy" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

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
  function_name = aws_lambda_function.lambda.arn
  principal     = "apigateway.${data.aws_partition.current.dns_suffix}"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.confirm_subscription"
  source_code_hash = filebase64sha256("test-fixtures/lambda_confirm_sns.zip")
  runtime          = "python3.6"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = [aws_api_gateway_integration_response.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "acctest"
}

resource "aws_sns_topic_subscription" "test" {
  depends_on             = [aws_lambda_permission.apigw_lambda]
  topic_arn              = aws_sns_topic.test.arn
  protocol               = "https"
  endpoint               = aws_api_gateway_deployment.test.invoke_url
  endpoint_auto_confirms = true
}
`, rName)
}

func testAccAWSSNSTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api" "test" {
  name        = %[1]q
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_rest_api.test.root_resource_id
  http_method   = "POST"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.test.id
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = aws_api_gateway_rest_api.test.id
  resource_id             = aws_api_gateway_rest_api.test.root_resource_id
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = [aws_api_gateway_integration.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "iam_for_lambda" {
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

resource "aws_iam_role_policy" "policy" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

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
  function_name = aws_lambda_function.lambda.arn
  principal     = "apigateway.${data.aws_partition.current.dns_suffix}"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.confirm_subscription"
  source_code_hash = filebase64sha256("test-fixtures/lambda_confirm_sns.zip")
  runtime          = "python3.6"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = [aws_api_gateway_integration_response.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "acctest"
}

resource "aws_iam_role" "invocation_role" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "%[1]s-2"
  role = aws_iam_role.invocation_role.id

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
  name                   = %[1]q
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}

resource "aws_lambda_function" "authorizer" {
  filename         = "test-fixtures/lambda_basic_authorizer.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambda_basic_authorizer.zip")
  function_name    = "%[1]s-2"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.authenticate"
  runtime          = "nodejs12.x"

  environment {
    variables = {
      AUTH_USER = "davematthews"
      AUTH_PASS = "granny"
    }
  }
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.WWW-Authenticate" = "'Basic'"
  }
}

resource "aws_sns_topic_subscription" "test" {
  depends_on             = [aws_lambda_permission.apigw_lambda]
  topic_arn              = aws_sns_topic.test.arn
  protocol               = "https"
  endpoint               = replace(aws_api_gateway_deployment.test.invoke_url, "https://", "https://davematthews:granny@")
  endpoint_auto_confirms = true
}
`, rName)
}

func testAccAWSSNSTopicSubscriptionEmailConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "email"
  endpoint  = "invalid_email@example.com"
}
`, rName)
}

func testAccAWSSNSTopicSubscriptionConfig_firehose(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  endpoint              = aws_kinesis_firehose_delivery_stream.test_stream.arn
  protocol              = "firehose"
  topic_arn             = aws_sns_topic.test.arn
  subscription_role_arn = aws_iam_role.firehose_role.arn
}
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "firehose_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
  {
	"Action": "sts:AssumeRole",
	"Principal": {
	  "Service": ["sns.${data.aws_partition.current.dns_suffix}","firehose.${data.aws_partition.current.dns_suffix}"]
	},
	"Effect": "Allow",
	"Sid": ""
  }
]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName)
}
