package aws

import (
	"fmt"
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDynamoDbGlobalTable_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_dynamodb_global_table.test"
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDynamodbGlobalTable(t)
			testAccAlternateRegionPreCheck(t)
			testAccDynamoDBGlobalTablePreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsDynamoDbGlobalTableDestroy,
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
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "dynamodb", regexp.MustCompile("global-table/[a-z0-9-]+$")),
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
			testAccDynamoDBGlobalTablePreCheck(t)
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

// testAccDynamoDBGlobalTablePreCheck checks if aws_dynamodb_global_table (version 2017.11.29) can be used and skips test if not.
func testAccDynamoDBGlobalTablePreCheck(t *testing.T) {
	supportRegionsSort := []string{"ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "eu-central-1", "eu-west-1", "eu-west-2", "us-east-1", "us-east-2", "us-west-1", "us-west-2"}

	if testAccGetRegion() != supportRegionsSort[sort.SearchStrings(supportRegionsSort, testAccGetRegion())] {
		t.Skipf("skipping test; aws_dynamodb_global_table (DynamoDB v2017.11.29) not supported in region %s", testAccGetRegion())
	}
}

func testAccDynamoDbGlobalTableConfig_basic(tableName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {
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
  depends_on = [aws_dynamodb_table.test]

  name = "%s"

  replica {
    region_name = data.aws_region.current.name
  }
}
`, tableName, tableName)
}

func testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName string) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
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
  provider = "awsalternate"

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
	return testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName) + `
resource "aws_dynamodb_global_table" "test" {
  name = aws_dynamodb_table.test.name

  replica {
    region_name = data.aws_region.current.name
  }
}
`
}

func testAccDynamoDbGlobalTableConfig_multipleRegions2(tableName string) string {
	return testAccDynamoDbGlobalTableConfig_multipleRegions_dynamodb_tables(tableName) + `
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
`
}

func testAccDynamoDbGlobalTableConfig_invalidName(tableName string) string {
	return composeConfig(testAccAlternateRegionProviderConfig(), fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_global_table" "test" {
  name = "%s"

  replica {
    region_name = data.aws_region.alternate.name
  }
}
`, tableName))
}
