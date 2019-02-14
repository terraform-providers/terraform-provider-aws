package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDynamoDbGlobalTable_basic(t *testing.T) {
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDynamoDbGlobalTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDynamoDbGlobalTableConfig_invalidName(acctest.RandString(2)),
				ExpectError: regexp.MustCompile("name length must be between 3 and 255 characters"),
			},
			{
				Config:      testAccDynamoDbGlobalTableConfig_invalidName(acctest.RandString(256)),
				ExpectError: regexp.MustCompile("name length must be between 3 and 255 characters"),
			},
			{
				Config:      testAccDynamoDbGlobalTableConfig_invalidName("!!!!"),
				ExpectError: regexp.MustCompile("name must only include alphanumeric, underscore, period, or hyphen characters"),
			},
			{
				Config: testAccDynamoDbGlobalTableConfig_basic(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile("^arn:aws:dynamodb::[0-9]{12}:global-table/[a-z0-9-]+$")),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbGlobalTable_multipleRegions(t *testing.T) {
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDynamoDbGlobalTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replica.2896117718.region_name", "us-east-1"),
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile("^arn:aws:dynamodb::[0-9]{12}:global-table/[a-z0-9-]+$")),
				),
			},
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions2(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replica.2896117718.region_name", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "replica.2276617237.region_name", "us-east-2"),
				),
			},
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions3(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replica.2896117718.region_name", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "replica.3965887460.region_name", "us-west-2"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbGlobalTable_import(t *testing.T) {
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamoDbGlobalTableConfig_basic(tableName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDynamoDbGlobalTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_global_table" {
			continue
		}

		input := &dynamodb.DescribeGlobalTableInput{
			GlobalTableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeGlobalTable(input)
		if err != nil {
			if isAWSErr(err, dynamodb.ErrCodeGlobalTableNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected DynamoDB Global Table to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsDynamoDbGlobalTableExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDynamoDbGlobalTableConfig_basic(tableName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
  current = true
}

resource "aws_dynamodb_table" "test" {
  hash_key         = "myAttribute"
  name             = "%s"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_global_table" "test" {
  depends_on = ["aws_dynamodb_table.test"]

  name = "%s"

  replica {
    region_name = "${data.aws_region.current.name}"
  }
}
`, tableName, tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName string) string {
	return fmt.Sprintf(`
provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

provider "aws" {
  alias  = "us-east-2"
  region = "us-east-2"
}

provider "aws" {
  alias  = "us-west-2"
  region = "us-west-2"
}

resource "aws_dynamodb_table" "us-east-1" {
  provider = "aws.us-east-1"

  hash_key         = "myAttribute"
  name             = "%s"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_table" "us-east-2" {
  provider = "aws.us-east-2"

  hash_key         = "myAttribute"
  name             = "%s"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_table" "us-west-2" {
  provider = "aws.us-west-2"

  hash_key         = "myAttribute"
  name             = "%s"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}`, tableName, tableName, tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName string) string {
	return fmt.Sprintf(`
%s

resource "aws_dynamodb_global_table" "test" {
  depends_on = ["aws_dynamodb_table.us-east-1"]
  provider = "aws.us-east-1"

  name = "%s"

  replica {
    region_name = "us-east-1"
  }
}`, testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName), tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions2(tableName string) string {
	return fmt.Sprintf(`
%s

resource "aws_dynamodb_global_table" "test" {
  depends_on = ["aws_dynamodb_table.us-east-1", "aws_dynamodb_table.us-east-2"]
  provider = "aws.us-east-1"

  name = "%s"

  replica {
    region_name = "us-east-1"
  }

  replica {
    region_name = "us-east-2"
  }
}`, testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName), tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions3(tableName string) string {
	return fmt.Sprintf(`
%s

resource "aws_dynamodb_global_table" "test" {
  depends_on = ["aws_dynamodb_table.us-east-1", "aws_dynamodb_table.us-west-2"]
  provider = "aws.us-east-1"

  name = "%s"

  replica {
    region_name = "us-east-1"
  }

  replica {
    region_name = "us-west-2"
  }
}`, testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName), tableName)
}

func testAccDynamoDbGlobalTableConfig_invalidName(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_table" "test" {
  name = "%s"

  replica {
    region_name = "us-east-1"
  }
}`, tableName)
}
