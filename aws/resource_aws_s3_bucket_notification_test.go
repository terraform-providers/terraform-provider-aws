package aws

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestAccAWSS3BucketNotification_importBasic(t *testing.T) {
	rString := acctest.RandString(8)

	topicName := fmt.Sprintf("tf-acc-topic-s3-b-n-import-%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-n-import-%s", rString)
	queueName := fmt.Sprintf("tf-acc-queue-s3-b-n-import-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-s3-b-n-import-%s", rString)
	lambdaFuncName := fmt.Sprintf("tf-acc-lambda-func-s3-b-n-import-%s", rString)

	resourceName := "aws_s3_bucket_notification.notification"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithTopicNotification(topicName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithQueueNotification(queueName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithLambdaNotification(roleName, lambdaFuncName, bucketName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSS3BucketNotification_basic(t *testing.T) {
	rString := acctest.RandString(8)

	topicName := fmt.Sprintf("tf-acc-topic-s3-b-notification-%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-notification-%s", rString)
	queueName := fmt.Sprintf("tf-acc-queue-s3-b-notification-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-s3-b-notification-%s", rString)
	lambdaFuncName := fmt.Sprintf("tf-acc-lambda-func-s3-b-notification-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithTopicNotification(topicName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketTopicNotification(
						"aws_s3_bucket.bucket",
						"notification-sns1",
						"aws_sns_topic.topic",
						[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:Delete"},
						&s3.KeyFilter{
							FilterRules: []*s3.FilterRule{
								{
									Name:  aws.String("Prefix"),
									Value: aws.String("tf-acc-test/"),
								},
								{
									Name:  aws.String("Suffix"),
									Value: aws.String(".txt"),
								},
							},
						},
					),
					testAccCheckAWSS3BucketTopicNotification(
						"aws_s3_bucket.bucket",
						"notification-sns2",
						"aws_sns_topic.topic",
						[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:Delete"},
						&s3.KeyFilter{
							FilterRules: []*s3.FilterRule{
								{
									Name:  aws.String("Suffix"),
									Value: aws.String(".log"),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithQueueNotification(queueName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketQueueNotification(
						"aws_s3_bucket.bucket",
						"notification-sqs",
						"aws_sqs_queue.queue",
						[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:Delete"},
						&s3.KeyFilter{
							FilterRules: []*s3.FilterRule{
								{
									Name:  aws.String("Prefix"),
									Value: aws.String("tf-acc-test/"),
								},
								{
									Name:  aws.String("Suffix"),
									Value: aws.String(".mp4"),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccAWSS3BucketConfigWithLambdaNotification(roleName, lambdaFuncName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketLambdaFunctionConfiguration(
						"aws_s3_bucket.bucket",
						"notification-lambda",
						"aws_lambda_function.func",
						[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:Delete"},
						&s3.KeyFilter{
							FilterRules: []*s3.FilterRule{
								{
									Name:  aws.String("Prefix"),
									Value: aws.String("tf-acc-test/"),
								},
								{
									Name:  aws.String("Suffix"),
									Value: aws.String(".png"),
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccAWSS3BucketNotification_withoutFilter(t *testing.T) {
	rString := acctest.RandString(8)

	topicName := fmt.Sprintf("tf-acc-topic-s3-b-notification-wo-f-%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-notification-wo-f-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketConfigWithTopicNotificationWithoutFilter(topicName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketTopicNotification(
						"aws_s3_bucket.bucket",
						"notification-sns1",
						"aws_sns_topic.topic",
						[]string{"s3:ObjectCreated:*", "s3:ObjectRemoved:Delete"},
						nil,
					),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketNotificationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_notification" {
			continue
		}
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			out, err := conn.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
				Bucket: aws.String(rs.Primary.ID),
			})
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucket" {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if len(out.TopicConfigurations) > 0 {
				return resource.RetryableError(fmt.Errorf("TopicConfigurations is exists: %v", out))
			}
			if len(out.LambdaFunctionConfigurations) > 0 {
				return resource.RetryableError(fmt.Errorf("LambdaFunctionConfigurations is exists: %v", out))
			}
			if len(out.QueueConfigurations) > 0 {
				return resource.RetryableError(fmt.Errorf("QueueConfigurations is exists: %v", out))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckAWSS3BucketTopicNotification(n, i, t string, events []string, filters *s3.KeyFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		topicArn := s.RootModule().Resources[t].Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			out, err := conn.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
				Bucket: aws.String(rs.Primary.ID),
			})

			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("GetBucketNotification error: %v", err))
			}

			eventSlice := sort.StringSlice(events)
			eventSlice.Sort()

			outputTopics := out.TopicConfigurations
			matched := false
			for _, outputTopic := range outputTopics {
				if *outputTopic.Id == i {
					matched = true

					if *outputTopic.TopicArn != topicArn {
						return resource.RetryableError(fmt.Errorf("bad topic arn, expected: %s, got %#v", topicArn, *outputTopic.TopicArn))
					}

					if filters != nil {
						if !reflect.DeepEqual(filters, outputTopic.Filter.Key) {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: %#v, got %#v", filters, outputTopic.Filter.Key))
						}
					} else {
						if outputTopic.Filter != nil {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: nil, got %#v", outputTopic.Filter))
						}
					}

					outputEventSlice := sort.StringSlice(aws.StringValueSlice(outputTopic.Events))
					outputEventSlice.Sort()
					if !reflect.DeepEqual(eventSlice, outputEventSlice) {
						return resource.RetryableError(fmt.Errorf("bad notification events, expected: %#v, got %#v", events, outputEventSlice))
					}
				}
			}

			if !matched {
				return resource.RetryableError(fmt.Errorf("No match topic configurations: %#v", out))
			}

			return nil
		})

		return err
	}
}

func testAccCheckAWSS3BucketQueueNotification(n, i, t string, events []string, filters *s3.KeyFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		queueArn := s.RootModule().Resources[t].Primary.Attributes["arn"]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			out, err := conn.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
				Bucket: aws.String(rs.Primary.ID),
			})

			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("GetBucketNotification error: %v", err))
			}

			eventSlice := sort.StringSlice(events)
			eventSlice.Sort()

			outputQueues := out.QueueConfigurations
			matched := false
			for _, outputQueue := range outputQueues {
				if *outputQueue.Id == i {
					matched = true

					if *outputQueue.QueueArn != queueArn {
						return resource.RetryableError(fmt.Errorf("bad queue arn, expected: %s, got %#v", queueArn, *outputQueue.QueueArn))
					}

					if filters != nil {
						if !reflect.DeepEqual(filters, outputQueue.Filter.Key) {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: %#v, got %#v", filters, outputQueue.Filter.Key))
						}
					} else {
						if outputQueue.Filter != nil {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: nil, got %#v", outputQueue.Filter))
						}
					}

					outputEventSlice := sort.StringSlice(aws.StringValueSlice(outputQueue.Events))
					outputEventSlice.Sort()
					if !reflect.DeepEqual(eventSlice, outputEventSlice) {
						return resource.RetryableError(fmt.Errorf("bad notification events, expected: %#v, got %#v", events, outputEventSlice))
					}
				}
			}

			if !matched {
				return resource.RetryableError(fmt.Errorf("No match queue configurations: %#v", out))
			}

			return nil
		})

		return err
	}
}

func testAccCheckAWSS3BucketLambdaFunctionConfiguration(n, i, t string, events []string, filters *s3.KeyFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		funcArn := s.RootModule().Resources[t].Primary.Attributes["arn"]
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			out, err := conn.GetBucketNotificationConfiguration(&s3.GetBucketNotificationConfigurationRequest{
				Bucket: aws.String(rs.Primary.ID),
			})

			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("GetBucketNotification error: %v", err))
			}

			eventSlice := sort.StringSlice(events)
			eventSlice.Sort()

			outputFunctions := out.LambdaFunctionConfigurations
			matched := false
			for _, outputFunc := range outputFunctions {
				if *outputFunc.Id == i {
					matched = true

					if *outputFunc.LambdaFunctionArn != funcArn {
						return resource.RetryableError(fmt.Errorf("bad lambda function arn, expected: %s, got %#v", funcArn, *outputFunc.LambdaFunctionArn))
					}

					if filters != nil {
						if !reflect.DeepEqual(filters, outputFunc.Filter.Key) {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: %#v, got %#v", filters, outputFunc.Filter.Key))
						}
					} else {
						if outputFunc.Filter != nil {
							return resource.RetryableError(fmt.Errorf("bad notification filters, expected: nil, got %#v", outputFunc.Filter))
						}
					}

					outputEventSlice := sort.StringSlice(aws.StringValueSlice(outputFunc.Events))
					outputEventSlice.Sort()
					if !reflect.DeepEqual(eventSlice, outputEventSlice) {
						return resource.RetryableError(fmt.Errorf("bad notification events, expected: %#v, got %#v", events, outputEventSlice))
					}
				}
			}

			if !matched {
				return resource.RetryableError(fmt.Errorf("No match lambda function configurations: %#v", out))
			}

			return nil
		})

		return err
	}
}

func testAccAWSS3BucketConfigWithTopicNotification(topicName, bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sns_topic" "topic" {
    name = "%s"
	policy = <<POLICY
{
	"Version":"2012-10-17",
	"Statement":[{
		"Sid": "",
		"Effect": "Allow",
		"Principal": {"AWS":"*"},
		"Action": "SNS:Publish",
		"Resource": "arn:${data.aws_partition.current.partition}:sns:*:*:%s",
		"Condition":{
			"ArnLike":{"aws:SourceArn":"${aws_s3_bucket.bucket.arn}"}
		}
	}]
}
POLICY
}

resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	acl = "public-read"
}

resource "aws_s3_bucket_notification" "notification" {
	bucket = "${aws_s3_bucket.bucket.id}"
	topic {
		id = "notification-sns1"
		topic_arn = "${aws_sns_topic.topic.arn}"
		events = [
		  "s3:ObjectCreated:*",
		  "s3:ObjectRemoved:Delete",
		]
		filter_prefix = "tf-acc-test/"
		filter_suffix = ".txt"
	}
	topic {
		id = "notification-sns2"
		topic_arn = "${aws_sns_topic.topic.arn}"
		events = [
		  "s3:ObjectCreated:*",
		  "s3:ObjectRemoved:Delete",
		]
		filter_suffix = ".log"
	}
}
`, topicName, topicName, bucketName)
}

func testAccAWSS3BucketConfigWithQueueNotification(queueName, bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sqs_queue" "queue" {
    name = "%s"
	policy = <<POLICY
{
	"Version":"2012-10-17",
	"Statement": [{
		"Effect":"Allow",
		"Principal":"*",
		"Action":"sqs:SendMessage",
		"Resource":"arn:${data.aws_partition.current.partition}:sqs:*:*:%s",
		"Condition":{
			"ArnEquals":{
				"aws:SourceArn":"${aws_s3_bucket.bucket.arn}"
			}
		}
	}]
}
POLICY
}

resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	acl = "public-read"
}

resource "aws_s3_bucket_notification" "notification" {
	bucket = "${aws_s3_bucket.bucket.id}"
	queue {
		id = "notification-sqs"
		queue_arn = "${aws_sqs_queue.queue.arn}"
		events = [
		  "s3:ObjectCreated:*",
		  "s3:ObjectRemoved:Delete",
		]
		filter_prefix = "tf-acc-test/"
		filter_suffix = ".mp4"
	}
}
`, queueName, queueName, bucketName)
}

func testAccAWSS3BucketConfigWithLambdaNotification(roleName, lambdaFuncName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
    name = "%s"
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

resource "aws_lambda_permission" "allow_bucket" {
    statement_id = "AllowExecutionFromS3Bucket"
    action = "lambda:InvokeFunction"
    function_name = "${aws_lambda_function.func.arn}"
    principal = "s3.amazonaws.com"
    source_arn = "${aws_s3_bucket.bucket.arn}"
}

resource "aws_lambda_function" "func" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
		runtime = "nodejs8.10"
}

resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	acl = "public-read"
}

resource "aws_s3_bucket_notification" "notification" {
	bucket = "${aws_s3_bucket.bucket.id}"
	lambda_function {
		id = "notification-lambda"
		lambda_function_arn = "${aws_lambda_function.func.arn}"
		events = [
		  "s3:ObjectCreated:*",
		  "s3:ObjectRemoved:Delete",
		]
		filter_prefix = "tf-acc-test/"
		filter_suffix = ".png"
	}
}
`, roleName, lambdaFuncName, bucketName)
}

func testAccAWSS3BucketConfigWithTopicNotificationWithoutFilter(topicName, bucketName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
resource "aws_sns_topic" "topic" {
    name = "%s"
	policy = <<POLICY
{
	"Version":"2012-10-17",
	"Statement":[{
		"Sid": "",
		"Effect": "Allow",
		"Principal": {"AWS":"*"},
		"Action": "SNS:Publish",
		"Resource": "arn:${data.aws_partition.current.partition}:sns:*:*:%s",
		"Condition":{
			"ArnLike":{"aws:SourceArn":"${aws_s3_bucket.bucket.arn}"}
		}
	}]
}
POLICY
}

resource "aws_s3_bucket" "bucket" {
	bucket = "%s"
	acl = "public-read"
}

resource "aws_s3_bucket_notification" "notification" {
	bucket = "${aws_s3_bucket.bucket.id}"
	topic {
		id = "notification-sns1"
		topic_arn = "${aws_sns_topic.topic.arn}"
		events = [
		  "s3:ObjectCreated:*",
		  "s3:ObjectRemoved:Delete",
		]
	}
}
`, topicName, topicName, bucketName)
}
