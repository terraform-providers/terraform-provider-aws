package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfsqs "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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

	err = conn.ListQueuesPages(input, func(page *sqs.ListQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queueUrl := range page.QueueUrls {
			r := resourceAwsSqsQueue()
			d := r.Data(nil)
			d.SetId(aws.StringValue(queueUrl))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SQS Queue sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing SQS Queues: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSQSQueue_basic(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", strconv.Itoa(tfsqs.DefaultQueueDelaySeconds)),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKmsDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", strconv.Itoa(tfsqs.DefaultQueueMaximumMessageSize)),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", strconv.Itoa(tfsqs.DefaultQueueMessageRetentionPeriod)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", strconv.Itoa(tfsqs.DefaultQueueReceiveMessageWaitTimeSeconds)),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "url", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", strconv.Itoa(tfsqs.DefaultQueueVisibilityTimeout)),
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

func TestAccAWSSQSQueue_disappears(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueue_Name_Generated(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueueConfigNameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
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

func TestAccAWSSQSQueue_Name_Generated_FIFOQueue(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueueConfigNameGeneratedFIFOQueue,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					naming.TestCheckResourceAttrNameWithSuffixGenerated(resourceName, "name", tfsqs.FifoQueueNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccAWSSQSQueue_NamePrefix(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueueConfigNamePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
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

func TestAccAWSSQSQueue_NamePrefix_FIFOQueue(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueueConfigNamePrefixFIFOQueue("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					naming.TestCheckResourceAttrNameWithSuffixFromPrefix(resourceName, "name", "tf-acc-test-prefix-", tfsqs.FifoQueueNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccAWSSQSQueue_Tags(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
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
				Config: testAccAWSSQSConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSQSConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_Update(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", strconv.Itoa(tfsqs.DefaultQueueDelaySeconds)),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKmsDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", strconv.Itoa(tfsqs.DefaultQueueMaximumMessageSize)),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", strconv.Itoa(tfsqs.DefaultQueueMessageRetentionPeriod)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", strconv.Itoa(tfsqs.DefaultQueueReceiveMessageWaitTimeSeconds)),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", strconv.Itoa(tfsqs.DefaultQueueVisibilityTimeout)),
				),
			},
			{
				Config: testAccAWSSQSConfigUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sqs", rName),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "false"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKmsDataKeyReusePeriodSeconds)),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", "2048"),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "60"),
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

func TestAccAWSSQSQueue_Policy(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckAWSSQSQueuePolicyAttribute(&queueAttributes, rName),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "90"),
					resource.TestCheckResourceAttr(resourceName, "max_message_size", "2048"),
					resource.TestCheckResourceAttr(resourceName, "message_retention_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "receive_wait_time_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "60"),
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

func TestAccAWSSQSQueue_RecentlyDeleted(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSSQSConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_RedrivePolicy(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigRedrivePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "delay_seconds", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "redrive_policy"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "300"),
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

func TestAccAWSSQSQueue_FIFOQueue(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", acctest.RandomWithPrefix("tf-acc-test"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigFIFOQueue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "queue"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perQueue"),
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

func TestAccAWSSQSQueue_FIFOQueue_ExpectNameError(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSQSConfigFIFOQueue(rName),
				ExpectError: regexp.MustCompile(`invalid queue name:`),
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFOQueue_ContentBasedDeduplication(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", acctest.RandomWithPrefix("tf-acc-test"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigFIFOQueueContentBasedDeduplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
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

func TestAccAWSSQSQueue_FIFOQueue_HighThroughputMode(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := fmt.Sprintf("%s.fifo", acctest.RandomWithPrefix("tf-acc-test"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigFIFOQueueHighThroughputMode(rName, "null", "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "queue"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perQueue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSQSConfigFIFOQueueHighThroughputMode(rName, "messageGroup", "perMessageGroupId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "deduplication_scope", "messageGroup"),
					resource.TestCheckResourceAttr(resourceName, "fifo_queue", "true"),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_limit", "perMessageGroupId"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_StandardQueue_ExpectContentBasedDeduplicationError(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSQSConfigStandardQueueExpectContentBasedDeduplicationError(rName),
				ExpectError: regexp.MustCompile(`content-based deduplication can only be set for FIFO queue`),
			},
		},
	})
}

func TestAccAWSSQSQueue_Encryption(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigEncryption(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sqs"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSQSConfigEncryption(rName, "3600"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sqs"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_ZeroVisibilityTimeoutSeconds(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigZeroVisibilityTimeoutSeconds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "0"),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/19786.
func TestAccAWSSQSQueue_DefaultKmsDataKeyReusePeriodSeconds(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigDefaultKmsDataKeyReusePeriodSeconds(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(resourceName, &queueAttributes),
					resource.TestCheckResourceAttr(resourceName, "kms_data_key_reuse_period_seconds", strconv.Itoa(tfsqs.DefaultQueueKmsDataKeyReusePeriodSeconds)),
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

func testAccCheckAWSSQSQueuePolicyAttribute(queueAttributes *map[string]string, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedPolicyText := fmt.Sprintf(
			`{
"Version": "2012-10-17",
"Id": "sqspolicy",
"Statement":[{
  "Sid": "Stmt1451501026839",
  "Effect": "Allow",
  "Principal":"*",
  "Action":"sqs:SendMessage",
  "Resource":"arn:%[1]s:sqs:%[2]s:%[3]s:%[4]s",
  "Condition":{
    "ArnEquals":{"aws:SourceArn":"arn:%[1]s:sns:%[2]s:%[3]s:%[5]s"}
  }
}]
             }`,
			testAccGetPartition(), testAccGetRegion(), testAccGetAccountID(), rName, rName)

		var actualPolicyText string
		for key, value := range *queueAttributes {
			if key == sqs.QueueAttributeNamePolicy {
				actualPolicyText = value
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n", expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSSQSQueueExists(resourceName string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SQS Queue URL is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sqsconn

		output, err := finder.QueueAttributesByURL(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccCheckAWSSQSQueueDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sqsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sqs_queue" {
			continue
		}

		_, err := finder.QueueAttributesByURL(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SQS Queue %s still exists", rs.Primary.ID)
	}

	return nil
}

const testAccAWSSQSQueueConfigNameGenerated = `
resource "aws_sqs_queue" "test" {}
`

const testAccAWSSQSQueueConfigNameGeneratedFIFOQueue = `
resource "aws_sqs_queue" "test" {
  fifo_queue = true
}
`

func testAccAWSSQSConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSSQSQueueConfigNamePrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name_prefix = %[1]q
}
`, prefix)
}

func testAccAWSSQSQueueConfigNamePrefixFIFOQueue(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name_prefix = %[1]q
  fifo_queue  = true
}
`, prefix)
}

func testAccAWSSQSConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSQSConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSQSConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60
}
`, rName)
}

func testAccAWSSQSConfigPolicy(rName string) string {
	return fmt.Sprintf(`
locals {
  queue_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_sqs_queue" "test" {
  name                       = local.queue_name
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
      "Resource": "arn:${data.aws_partition.current.partition}:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${local.queue_name}",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "${aws_sns_topic.test.arn}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccAWSSQSConfigRedrivePolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = "%[1]s-1"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  redrive_policy = <<EOF
{
  "maxReceiveCount": 3,
  "deadLetterTargetArn": "${aws_sqs_queue.dlq.arn}"
}
EOF
}

resource "aws_sqs_queue" "dlq" {
  name = "%[1]s-2"
}
`, rName)
}

func testAccAWSSQSConfigFIFOQueue(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name       = %[1]q
  fifo_queue = true
}
`, rName)
}

func testAccAWSSQSConfigFIFOQueueContentBasedDeduplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                        = %[1]q
  fifo_queue                  = true
  content_based_deduplication = true
}
`, rName)
}

func testAccAWSSQSConfigFIFOQueueHighThroughputMode(rName, deduplicationScope, fifoThroughputLimit string) string {
	if deduplicationScope != "null" {
		deduplicationScope = strconv.Quote(deduplicationScope)
	}

	if fifoThroughputLimit != "null" {
		fifoThroughputLimit = strconv.Quote(fifoThroughputLimit)
	}

	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name       = %[1]q
  fifo_queue = true

  deduplication_scope   = %[2]s
  fifo_throughput_limit = %[3]s
}
`, rName, deduplicationScope, fifoThroughputLimit)
}

func testAccAWSSQSConfigStandardQueueExpectContentBasedDeduplicationError(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                        = %[1]q
  content_based_deduplication = true
}
`, rName)
}

func testAccAWSSQSConfigEncryption(rName, kmsDataKeyReusePeriodSeconds string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                              = %[1]q
  kms_master_key_id                 = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = %[2]s
}
`, rName, kmsDataKeyReusePeriodSeconds)
}

func testAccAWSSQSConfigZeroVisibilityTimeoutSeconds(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                       = %[1]q
  visibility_timeout_seconds = 0
}
`, rName)
}

func testAccAWSSQSConfigDefaultKmsDataKeyReusePeriodSeconds(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name                              = %[1]q
  kms_data_key_reuse_period_seconds = 300
}
`, rName)
}
