package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesis/finder"
)

func TestAccAWSKinesisStreamConsumer_basic(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kinesis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKinesisStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSKinesisStreamConsumerExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kinesis", regexp.MustCompile(fmt.Sprintf("stream/%[1]s/consumer/%[1]s", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "stream_arn", streamName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
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

func TestAccAWSKinesisStreamConsumer_disappears(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kinesis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKinesisStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSKinesisStreamConsumerExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisStreamConsumer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisStreamConsumer_MaxConcurrentConsumers(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kinesis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				// Test creation of max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccAWSKinesisStreamConsumerConfig_multiple(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.0", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.1", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.2", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.3", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.4", resourceName)),
				),
			},
		},
	})
}

func TestAccAWSKinesisStreamConsumer_ExceedMaxConcurrentConsumers(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kinesis.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKinesisStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				// Test creation of more than the max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccAWSKinesisStreamConsumerConfig_multiple(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.0", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.1", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.2", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.3", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.4", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.5", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.6", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.7", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.8", resourceName)),
					testAccAWSKinesisStreamConsumerExists(fmt.Sprintf("%s.9", resourceName)),
				),
			},
		},
	})
}

func testAccCheckAWSKinesisStreamConsumerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_stream_consumer" {
			continue
		}

		consumer, err := finder.StreamConsumerByARN(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", rs.Primary.ID, err)
		}

		if consumer != nil {
			return fmt.Errorf("Kinesis Stream Consumer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSKinesisStreamConsumerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisconn

		consumer, err := finder.StreamConsumerByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", rs.Primary.ID, err)
		}

		if consumer == nil {
			return fmt.Errorf("Kinesis Stream Consumer (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSKinesisStreamConsumerBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %q
  shard_count = 2
}
`, rName)
}

func testAccAWSKinesisStreamConsumerConfig_basic(rName string) string {
	return composeConfig(
		testAccAWSKinesisStreamConsumerBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  name       = %q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}

func testAccAWSKinesisStreamConsumerConfig_multiple(rName string, count int) string {
	return composeConfig(
		testAccAWSKinesisStreamConsumerBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  count      = %d
  name       = "%s-${count.index}"
  stream_arn = aws_kinesis_stream.test.arn
}
`, count, rName))
}
