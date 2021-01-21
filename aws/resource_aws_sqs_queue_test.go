package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func init() {
	resource.AddTestSweepers("aws_sqs_queue", &resource.Sweeper{
		Name: "aws_sqs_queue",
		F:    testSweepSqsQueues,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_cloudwatch_event_rule",
			"aws_elastic_beanstalk_environment",
			"aws_iot_topic_rule",
			"aws_lambda_function",
			"aws_s3_bucket",
			"aws_sns_topic",
		},
	})
}

func testSweepSqsQueues(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).sqsconn
	input := &sqs.ListQueuesInput{}
	var sweeperErrs *multierror.Error

	output, err := conn.ListQueues(input)
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SQS Queues sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SQS Queues: %w", err))
		return sweeperErrs
	}

	for _, queueUrl := range output.QueueUrls {
		url := aws.StringValue(queueUrl)

		log.Printf("[INFO] Deleting SQS Queue: %s", url)
		_, err := conn.DeleteQueue(&sqs.DeleteQueueInput{
			QueueUrl: aws.String(url),
		})
		if isAWSErr(err, sqs.ErrCodeQueueDoesNotExist, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting SQS Queue (%s): %w", url, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSQSQueue_basic(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
			{
				Config: testAccAWSSQSConfigWithOverrides(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueOverrideAttributes(&queueAttributes),
				),
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_tags(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithTags(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
			{
				Config: testAccAWSSQSConfigWithTagsChanged(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_namePrefix(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	prefix := "acctest-sqs-queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithNamePrefix(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^acctest-sqs-queue`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_namePrefix_fifo(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	prefix := "acctest-sqs-queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSFifoConfigWithNamePrefix(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^acctest-sqs-queue.*\.fifo$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_policy(t *testing.T) {
	var queueAttributes map[string]*string

	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	topicName := fmt.Sprintf("sns-topic-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfig_PolicyFormat(topicName, queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.test-email-events", &queueAttributes),
					testAccCheckAWSSQSQueuePolicyAttribute(&queueAttributes, topicName, queueName),
				),
			},
			{
				ResourceName:      "aws_sqs_queue.test-email-events",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_queueDeletedRecently(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
				),
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
				),
				Taint: []string{resourceName},
			},
		},
	})
}

func TestAccAWSSQSQueue_redrivePolicy(t *testing.T) {
	var queueAttributes map[string]*string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithRedrive(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.my_dead_letter_queue", &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
				),
			},
			{
				ResourceName:      "aws_sqs_queue.my_dead_letter_queue",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

// Tests formatting and compacting of Policy, Redrive json
func TestAccAWSSQSQueue_Policybasic(t *testing.T) {
	var queueAttributes map[string]*string

	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	topicName := fmt.Sprintf("sns-topic-%s", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfig_PolicyFormat(topicName, queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.test-email-events", &queueAttributes),
					testAccCheckAWSSQSQueueOverrideAttributes(&queueAttributes),
				),
			},
			{
				ResourceName:      "aws_sqs_queue.test-email-events",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFO(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithFIFO(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFOExpectNameError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSQSConfigWithFIFOExpectError(acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Error validating the FIFO queue name`),
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFOWithContentBasedDeduplication(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithFIFOContentBasedDeduplication(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func TestAccAWSSQSQueue_ExpectContentBasedDeduplicationError(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccExpectContentBasedDeduplicationError(acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Content based deduplication can only be set with FIFO queues`),
			},
		},
	})
}

func TestAccAWSSQSQueue_Encryption(t *testing.T) {
	var queueAttributes map[string]*string

	resourceName := "aws_sqs_queue.queue"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithEncryption(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sqs"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"name_prefix",
				},
			},
		},
	})
}

func testAccCheckAWSSQSQueueDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sqsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sqs_queue" {
			continue
		}

		// Check if queue exists by checking for its attributes
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(rs.Primary.ID),
		}
		err := resource.Retry(15*time.Second, func() *resource.RetryError {
			_, err := conn.GetQueueAttributes(params)
			if err != nil {
				if isAWSErr(err, sqs.ErrCodeQueueDoesNotExist, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			return resource.RetryableError(fmt.Errorf("Queue %s still exists. Failing!", rs.Primary.ID))
		})
		if err != nil {
			return err
		}
	}

	return nil
}
func testAccCheckAWSSQSQueuePolicyAttribute(queueAttributes *map[string]*string, topicName, queueName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		accountID := testAccProvider.Meta().(*AWSClient).accountid

		expectedPolicyFormat := `{"Version": "2012-10-17","Id": "sqspolicy","Statement":[{"Sid": "Stmt1451501026839","Effect": "Allow","Principal":"*","Action":"sqs:SendMessage","Resource":"arn:%[1]s:sqs:%[2]s:%[3]s:%[4]s","Condition":{"ArnEquals":{"aws:SourceArn":"arn:%[1]s:sns:%[2]s:%[3]s:%[5]s"}}}]}`
		expectedPolicyText := fmt.Sprintf(expectedPolicyFormat, testAccGetPartition(), testAccGetRegion(), accountID, topicName, queueName)

		var actualPolicyText string
		for key, valuePointer := range *queueAttributes {
			if key == "Policy" {
				actualPolicyText = aws.StringValue(valuePointer)
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSSQSQueueExists(resourceName string, queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).sqsconn

		input := &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(rs.Primary.ID),
			AttributeNames: []*string{aws.String("All")},
		}
		output, err := conn.GetQueueAttributes(input)

		if err != nil {
			return err
		}

		*queueAttributes = output.Attributes

		return nil
	}
}

func testAccCheckAWSSQSQueueDefaultAttributes(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if attributes are defaults
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "30" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 30", value)
			}

			if key == "MessageRetentionPeriod" && value != "345600" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 345600", value)
			}

			if key == "MaximumMessageSize" && value != "262144" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 262144", value)
			}

			if key == "DelaySeconds" && value != "0" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 0", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "0" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 0", value)
			}
		}

		return nil
	}
}

func testAccCheckAWSSQSQueueOverrideAttributes(queueAttributes *map[string]*string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// checking if attributes match our overrides
		for key, valuePointer := range *queueAttributes {
			value := aws.StringValue(valuePointer)
			if key == "VisibilityTimeout" && value != "60" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 60", value)
			}

			if key == "MessageRetentionPeriod" && value != "86400" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 86400", value)
			}

			if key == "MaximumMessageSize" && value != "2048" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 2048", value)
			}

			if key == "DelaySeconds" && value != "90" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 90", value)
			}

			if key == "ReceiveMessageWaitTimeSeconds" && value != "10" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 10", value)
			}
		}

		return nil
	}
}

func testAccAWSSQSConfigWithDefaults(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"
}
`, r)
}

func testAccAWSSQSConfigWithNamePrefix(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name_prefix = "%s"
}
`, r)
}

func testAccAWSSQSFifoConfigWithNamePrefix(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name_prefix = "%s"
  fifo_queue  = true
}
`, r)
}

func testAccAWSSQSConfigWithOverrides(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                       = "%s"
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60
}
`, r)
}

func testAccAWSSQSConfigWithRedrive(name string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "my_queue" {
  name                       = "tftestqueuq-%[1]s"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  redrive_policy = <<EOF
{
  "maxReceiveCount": 3,
  "deadLetterTargetArn": "${aws_sqs_queue.my_dead_letter_queue.arn}"
}
EOF
}

resource "aws_sqs_queue" "my_dead_letter_queue" {
  name = "tfotherqueuq-%[1]s"
}
`, name)
}

func testAccAWSSQSConfig_PolicyFormat(queue, topic string) string {
	return fmt.Sprintf(`
variable "sns_name" {
  default = "%s"
}

variable "sqs_name" {
  default = "%s"
}

resource "aws_sns_topic" "test_topic" {
  name = var.sns_name
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_sqs_queue" "test-email-events" {
  name                       = var.sqs_name
  depends_on                 = [aws_sns_topic.test_topic]
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement": [
    {
      "Sid": "Stmt1451501026839",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${var.sqs_name}",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${var.sns_name}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_sns_topic_subscription" "test_queue_target" {
  topic_arn = aws_sns_topic.test_topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test-email-events.arn
}
`, topic, queue)
}

func testAccAWSSQSConfigWithFIFO(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name       = "%s.fifo"
  fifo_queue = true
}
`, queue)
}

func testAccAWSSQSConfigWithFIFOContentBasedDeduplication(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                        = "%s.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
`, queue)
}

func testAccAWSSQSConfigWithFIFOExpectError(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name       = "%s"
  fifo_queue = true
}
`, queue)
}

func testAccExpectContentBasedDeduplicationError(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                        = "%s"
  content_based_deduplication = true
}
`, queue)
}

func testAccAWSSQSConfigWithEncryption(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                              = "%s"
  kms_master_key_id                 = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = 300
}
`, queue)
}

func testAccAWSSQSConfigWithTags(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, r)
}

func testAccAWSSQSConfigWithTagsChanged(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"

  tags = {
    Usage = "changed"
  }
}
`, r)
}
