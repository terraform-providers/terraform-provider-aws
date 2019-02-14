package aws

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSKinesisStream_basic(t *testing.T) {
	var stream kinesis.StreamDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_createMultipleConcurrentStreams(t *testing.T) {
	var stream kinesis.StreamDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfigConcurrent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.0", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.1", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.2", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.3", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.4", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.5", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.6", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.7", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.8", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.9", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.10", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.11", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.12", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.13", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.14", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.15", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.16", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.17", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.18", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream.19", &stream),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_encryptionWithoutKmsKeyThrowsError(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKinesisStreamConfigWithEncryptionAndNoKmsKey(rInt),
				ExpectError: regexp.MustCompile("KMS Key Id required when setting encryption_type is not set as NONE"),
			},
		},
	})
}

func TestAccAWSKinesisStream_encryption(t *testing.T) {
	var stream kinesis.StreamDescription
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfigWithEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "encryption_type", "KMS"),
				),
			},
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "encryption_type", "NONE"),
				),
			},
			{
				Config: testAccKinesisStreamConfigWithEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "encryption_type", "KMS"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_importBasic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_kinesis_stream.test_stream"
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     streamName,
			},
		},
	})
}

func TestAccAWSKinesisStream_shardCount(t *testing.T) {
	var stream kinesis.StreamDescription
	var updatedStream kinesis.StreamDescription

	testCheckStreamNotDestroyed := func() resource.TestCheckFunc {
		return func(*terraform.State) error {
			if *stream.StreamCreationTimestamp != *updatedStream.StreamCreationTimestamp {
				return fmt.Errorf("Creation timestamps dont match, stream was recreated")
			}
			return nil
		}
	}

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "shard_count", "2"),
				),
			},

			{
				Config: testAccKinesisStreamConfigUpdateShardCount(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &updatedStream),
					testAccCheckAWSKinesisStreamAttributes(&updatedStream),
					testCheckStreamNotDestroyed(),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "shard_count", "4"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_retentionPeriod(t *testing.T) {
	var stream kinesis.StreamDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "retention_period", "24"),
				),
			},

			{
				Config: testAccKinesisStreamConfigUpdateRetentionPeriod(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "retention_period", "100"),
				),
			},

			{
				Config: testAccKinesisStreamConfigDecreaseRetentionPeriod(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "retention_period", "28"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_shardLevelMetrics(t *testing.T) {
	var stream kinesis.StreamDescription

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckNoResourceAttr(
						"aws_kinesis_stream.test_stream", "shard_level_metrics"),
				),
			},

			{
				Config: testAccKinesisStreamConfigAllShardLevelMetrics(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "shard_level_metrics.#", "7"),
				),
			},

			{
				Config: testAccKinesisStreamConfigSingleShardLevelMetric(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						"aws_kinesis_stream.test_stream", "shard_level_metrics.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_Tags(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig_Tags(rInt, 21),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "21"),
				),
			},
			{
				Config: testAccKinesisStreamConfig_Tags(rInt, 9),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "9"),
				),
			},
		},
	})
}

func testAccCheckKinesisStreamExists(n string, stream *kinesis.StreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisconn
		describeOpts := &kinesis.DescribeStreamInput{
			StreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeStream(describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.StreamDescription

		return nil
	}
}

func testAccCheckAWSKinesisStreamAttributes(stream *kinesis.StreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.StreamName, "terraform-kinesis-test") {
			return fmt.Errorf("Bad Stream name: %s", *stream.StreamName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_stream" {
				continue
			}
			if *stream.StreamARN != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Stream ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *stream.StreamARN)
			}
			shard_count := strconv.Itoa(len(flattenShards(openShards(stream.Shards))))
			if shard_count != rs.Primary.Attributes["shard_count"] {
				return fmt.Errorf("Bad Stream Shard Count\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["shard_count"], shard_count)
			}
		}
		return nil
	}
}

func testAccCheckKinesisStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_stream" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).kinesisconn
		describeOpts := &kinesis.DescribeStreamInput{
			StreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeStream(describeOpts)
		if err == nil {
			if resp.StreamDescription != nil && *resp.StreamDescription.StreamStatus != "DELETING" {
				return fmt.Errorf("Error: Stream still exists")
			}
		}

		return nil

	}

	return nil
}

func testAccKinesisStreamConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigConcurrent(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
        count = 20
	name = "terraform-kinesis-test-%d-${count.index}"
	shard_count = 2
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigWithEncryptionAndNoKmsKey(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	encryption_type = "KMS"
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigWithEncryption(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	encryption_type = "KMS"
	kms_key_id = "${aws_kms_key.foo.id}"
	tags = {
		Name = "tf-test"
	}
}

resource "aws_kms_key" "foo" {
    description = "Kinesis Stream SSE AccTests %d"
    deletion_window_in_days = 7
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

`, rInt, rInt)
}

func testAccKinesisStreamConfigUpdateShardCount(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 4
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigUpdateRetentionPeriod(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	retention_period = 100
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigDecreaseRetentionPeriod(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	retention_period = 28
	tags = {
		Name = "tf-test"
	}
}`, rInt)
}

func testAccKinesisStreamConfigAllShardLevelMetrics(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	tags = {
		Name = "tf-test"
	}
	shard_level_metrics = [
		"IncomingBytes",
		"IncomingRecords",
		"OutgoingBytes",
		"OutgoingRecords",
		"WriteProvisionedThroughputExceeded",
		"ReadProvisionedThroughputExceeded",
		"IteratorAgeMilliseconds"
	]
}`, rInt)
}

func testAccKinesisStreamConfigSingleShardLevelMetric(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	tags = {
		Name = "tf-test"
	}
	shard_level_metrics = [
		"IncomingBytes"
	]
}`, rInt)
}

func testAccKinesisStreamConfig_Tags(rInt, tagCount int) string {
	var tagPairs string
	for i := 1; i <= tagCount; i++ {
		tagPairs = tagPairs + fmt.Sprintf("tag%d = \"tag%dvalue\"\n", i, i)
	}

	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
	name = "terraform-kinesis-test-%d"
	shard_count = 2
	tags = {
%s
	}
}`, rInt, tagPairs)
}
