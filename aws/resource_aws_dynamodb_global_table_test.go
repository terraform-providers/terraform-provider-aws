package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDynamoDbGlobalTable_basic(t *testing.T) {
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDynamodbGlobalTable(t) },
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbGlobalTable_multipleRegions(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDynamodbGlobalTable(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDynamoDbGlobalTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", tableName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
				),
			},
			{
				Config:            testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions2(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
			{
				Config: testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDynamoDbGlobalTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
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

func testAccPreCheckAWSDynamodbGlobalTable(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	input := &dynamodb.ListGlobalTablesInput{}

	_, err := conn.ListGlobalTables(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDynamoDbGlobalTableConfig_basic(tableName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

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
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "aws.alternate"
}

data "aws_region" "current" {}

resource "aws_dynamodb_table" "test" {
  hash_key         = "myAttribute"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_table" "alternate" {
  provider = "aws.alternate"

  hash_key         = "myAttribute"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}
`, tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions1(tableName string) string {
	return testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName) + fmt.Sprintf(`
resource "aws_dynamodb_global_table" "test" {
  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.current.name
  }
}
`)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions2(tableName string) string {
	return testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName) + fmt.Sprintf(`
resource "aws_dynamodb_global_table" "test" {
  depends_on = [aws_dynamodb_table.alternate]

  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name = data.aws_region.current.name
  }
}
`)
}

func testAccDynamoDbGlobalTableConfig_invalidName(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_global_table" "test" {
  name = "%s"

  replica {
    region_name = "us-east-1"
  }
}
`, tableName)
}
