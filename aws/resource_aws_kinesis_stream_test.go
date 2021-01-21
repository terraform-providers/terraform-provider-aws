package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_stream", &resource.Sweeper{
		Name: "aws_kinesis_stream",
		F:    testSweepKinesisStreams,
	})
}

func testSweepKinesisStreams(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).kinesisconn
	input := &kinesis.ListStreamsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListStreamsPages(input, func(page *kinesis.ListStreamsOutput, lastPage bool) bool {
		for _, streamNamePtr := range page.StreamNames {
			if streamNamePtr == nil {
				continue
			}

			streamName := aws.StringValue(streamNamePtr)
			input := &kinesis.DeleteStreamInput{
				EnforceConsumerDeletion: aws.Bool(false),
				StreamName:              streamNamePtr,
			}

			log.Printf("[INFO] Deleting Kinesis Stream: %s", streamName)

			_, err := conn.DeleteStream(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Kinesis Stream (%s): %w", streamName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Stream sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error listing Kinesis Streams: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSKinesisStream_basic(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"
	rInt := acctest.RandInt()
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccAWSKinesisStream_createMultipleConcurrentStreams(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"
	rInt := acctest.RandInt()
	streamName := fmt.Sprintf("terraform-kinesis-test-%d-0", rInt) // We can get away with just import testing one of them

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfigConcurrent(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.0", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.1", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.2", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.3", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.4", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.5", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.6", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.7", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.8", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.9", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.10", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.11", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.12", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.13", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.14", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.15", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.16", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.17", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.18", &stream),
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test.19", &stream),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
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
	resourceName := "aws_kinesis_stream.test"
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfigWithEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(
						resourceName, "encryption_type", "KMS"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(
						resourceName, "encryption_type", "NONE"),
				),
			},
			{
				Config: testAccKinesisStreamConfigWithEncryption(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(
						resourceName, "encryption_type", "KMS"),
				),
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
	resourceName := "aws_kinesis_stream.test"
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "shard_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccKinesisStreamConfigUpdateShardCount(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &updatedStream),
					testAccCheckAWSKinesisStreamAttributes(&updatedStream),
					testCheckStreamNotDestroyed(),
					resource.TestCheckResourceAttr(
						resourceName, "shard_count", "4"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_retentionPeriod(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"
	rInt := acctest.RandInt()
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "retention_period", "24"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccKinesisStreamConfigUpdateRetentionPeriod(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "retention_period", "8760"),
				),
			},

			{
				Config: testAccKinesisStreamConfigDecreaseRetentionPeriod(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "retention_period", "28"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_shardLevelMetrics(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"
	rInt := acctest.RandInt()
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckNoResourceAttr(
						resourceName, "shard_level_metrics"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccKinesisStreamConfigAllShardLevelMetrics(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "shard_level_metrics.#", "7"),
				),
			},

			{
				Config: testAccKinesisStreamConfigSingleShardLevelMetric(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					resource.TestCheckResourceAttr(
						resourceName, "shard_level_metrics.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisStream_enforceConsumerDeletion(t *testing.T) {
	var stream kinesis.StreamDescription
	resourceName := "aws_kinesis_stream.test"
	rInt := acctest.RandInt()
	streamName := fmt.Sprintf("terraform-kinesis-test-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfigWithEnforceConsumerDeletion(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisStreamAttributes(&stream),
					testAccAWSKinesisStreamRegisterStreamConsumer(&stream, fmt.Sprintf("tf-test-%d", rInt)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           streamName,
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
		},
	})
}

func TestAccAWSKinesisStream_Tags(t *testing.T) {
	var stream kinesis.StreamDescription
	rInt := acctest.RandInt()
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamConfig_Tags(rInt, 21),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckKinesisStreamTags(resourceName, 21),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("terraform-kinesis-test-%d", rInt),
				ImportStateVerifyIgnore: []string{"enforce_consumer_deletion"},
			},
			{
				Config: testAccKinesisStreamConfig_Tags(rInt, 9),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckKinesisStreamTags(resourceName, 9),
				),
			},
			{
				Config: testAccKinesisStreamConfig_Tags(rInt, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckKinesisStreamTags(resourceName, 50),
				),
			},
			{
				Config: testAccKinesisStreamConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					testAccCheckKinesisStreamTags(resourceName, 0),
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

func testAccAWSKinesisStreamRegisterStreamConsumer(stream *kinesis.StreamDescription, rStr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).kinesisconn

		if _, err := conn.RegisterStreamConsumer(&kinesis.RegisterStreamConsumerInput{
			ConsumerName: aws.String(rStr),
			StreamARN:    stream.StreamARN,
		}); err != nil {
			return err
		}

		return nil
	}
}
func testAccCheckKinesisStreamTags(n string, tagCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if err := resource.TestCheckResourceAttr(n, "tags.%", fmt.Sprintf("%d", tagCount))(s); err != nil {
			return err
		}

		for i := 0; i < tagCount; i++ {
			key := fmt.Sprintf("Key%0125d", i)
			value := fmt.Sprintf("Value%0251d", i)

			if err := resource.TestCheckResourceAttr(n, fmt.Sprintf("tags.%s", key), value)(s); err != nil {
				return err
			}
		}

		return nil
	}
}

func TestAccAWSKinesisStream_UpdateKmsKeyId(t *testing.T) {
	var stream kinesis.StreamDescription
	rInt := acctest.RandInt()
	resourceName := "aws_kinesis_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisStreamUpdateKmsKeyId(rInt, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.key.0", "id"),
				),
			},
			{
				Config: testAccKinesisStreamUpdateKmsKeyId(rInt, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", "aws_kms_key.key.1", "id"),
				),
			},
		},
	})
}

func testAccKinesisStreamConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "terraform-kinesis-test-%d"
  shard_count = 2
}
`, rInt)
}

func testAccKinesisStreamConfigConcurrent(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count       = 20
  name        = "terraform-kinesis-test-%d-${count.index}"
  shard_count = 2

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamConfigWithEncryptionAndNoKmsKey(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name            = "terraform-kinesis-test-%d"
  shard_count     = 2
  encryption_type = "KMS"

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamConfigWithEncryption(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name            = "terraform-kinesis-test-%d"
  shard_count     = 2
  encryption_type = "KMS"
  kms_key_id      = aws_kms_key.foo.id

  tags = {
    Name = "tf-test"
  }
}

resource "aws_kms_key" "foo" {
  description             = "Kinesis Stream SSE AccTests %d"
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
resource "aws_kinesis_stream" "test" {
  name        = "terraform-kinesis-test-%d"
  shard_count = 4

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamConfigUpdateRetentionPeriod(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = "terraform-kinesis-test-%d"
  shard_count      = 2
  retention_period = 8760

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamConfigDecreaseRetentionPeriod(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name             = "terraform-kinesis-test-%d"
  shard_count      = 2
  retention_period = 28

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamConfigAllShardLevelMetrics(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "terraform-kinesis-test-%d"
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
    "IteratorAgeMilliseconds",
  ]
}
`, rInt)
}

func testAccKinesisStreamConfigSingleShardLevelMetric(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "terraform-kinesis-test-%d"
  shard_count = 2

  tags = {
    Name = "tf-test"
  }

  shard_level_metrics = [
    "IncomingBytes",
  ]
}
`, rInt)
}

func testAccKinesisStreamConfig_Tags(rInt, tagCount int) string {
	// Tag limits:
	//  * Maximum number of tags per resource – 50
	//  * Maximum key length – 128 Unicode characters in UTF-8
	//  * Maximum value length – 256 Unicode characters in UTF-8
	tagPairs := make([]string, tagCount)
	for i := 0; i < tagCount; i++ {
		key := fmt.Sprintf("Key%0125d", i)
		value := fmt.Sprintf("Value%0251d", i)

		tagPairs[i] = fmt.Sprintf("%s = %q", key, value)
	}

	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = "terraform-kinesis-test-%[1]d"
  shard_count = 2

  tags = {
    %[2]s
  }
}
`, rInt, strings.Join(tagPairs, "\n"))
}

func testAccKinesisStreamConfigWithEnforceConsumerDeletion(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name                      = "terraform-kinesis-test-%d"
  shard_count               = 2
  enforce_consumer_deletion = true

  tags = {
    Name = "tf-test"
  }
}
`, rInt)
}

func testAccKinesisStreamUpdateKmsKeyId(rInt int, key int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "key" {
  count = 2

  description             = "KMS key ${count.index+1}"
  deletion_window_in_days = 10
}

resource "aws_kinesis_stream" "test" {
  name            = "test_stream-%[1]d"
  shard_count     = 1
  encryption_type = "KMS"
  kms_key_id      = aws_kms_key.key[%[2]d].id
}
`, rInt, key)
}
