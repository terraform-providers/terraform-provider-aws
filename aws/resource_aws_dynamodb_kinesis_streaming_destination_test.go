package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/dynamodb/finder"
)

func TestAccAwsDynamoDbKinesisStreamingDestination_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamoDbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbKinesisStreamingDestinationExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "stream_arn", "kinesis", regexp.MustCompile(fmt.Sprintf("stream/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName),
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

func TestAccAwsDynamoDbKinesisStreamingDestination_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamoDbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbKinesisStreamingDestinationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamoDbKinesisStreamingDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDynamoDbKinesisStreamingDestination_disappears_DynamoDbTable(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_dynamodb_kinesis_streaming_destination.test"
	tableResourceName := "aws_dynamodb_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSDynamoDbKinesisStreamingDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbKinesisStreamingDestinationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamoDbTable(), tableResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsDynamodbKinesisStreamingDestinationConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "hk"

  attribute {
    name = "hk"
    type = "S"
  }
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_dynamodb_kinesis_streaming_destination" "test" {
  table_name = aws_dynamodb_table.test.name
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName)
}

func testAccCheckDynamoDbKinesisStreamingDestinationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		tableName, streamArn, err := dynamoDbKinesisStreamingDestinationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		output, err := finder.DynamoDBKinesisDataStreamDestination(context.Background(), conn, streamArn, tableName)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("DynamoDB Kinesis Streaming Destination (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSDynamoDbKinesisStreamingDestinationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_kinesis_streaming_destination" {
			continue
		}

		tableName, streamArn, err := dynamoDbKinesisStreamingDestinationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := finder.DynamoDBKinesisDataStreamDestination(context.Background(), conn, streamArn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("DynamoDB Kinesis Streaming Destination (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
