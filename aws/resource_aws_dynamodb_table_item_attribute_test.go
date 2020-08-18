package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

const resourceName = "aws_dynamodb_table_item_attribute.test"

func TestAccAWSDynamoDbTableItemAttribute_basic(t *testing.T) {
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := acctest.RandString(8)
	attributeKey := acctest.RandString(8)
	attributeValue := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemAttributeDestroy(hashKey, ""),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemAttributeConfigBasic(tableName, hashKey, attributeKey, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", ""),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue),
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

func TestAccAWSDynamoDbTableItemAttribute_update(t *testing.T) {
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := acctest.RandString(8)
	attributeKey := acctest.RandString(8)
	attributeValue := acctest.RandString(8)
	attributeValue2 := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemAttributeDestroy(hashKey, ""),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemAttributeConfigBasic(tableName, hashKey, attributeKey, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", ""),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue),
				),
			},
			{
				Config: testAccAWSDynamoDbItemAttributeConfigBasic(tableName, hashKey, attributeKey, attributeValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", ""),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue2),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableItemAttribute_withRangeKey(t *testing.T) {
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := acctest.RandString(8)
	rangeKey := acctest.RandString(8)
	attributeKey := acctest.RandString(8)
	attributeValue := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemAttributeDestroy(hashKey, rangeKey),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemAttributeConfigWithRangeKey(tableName, hashKey, rangeKey, attributeKey, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", "rangeKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue),
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

func TestAccAWSDynamoDbTableItemAttribute_withRangeKey_update(t *testing.T) {
	tableName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	hashKey := acctest.RandString(8)
	rangeKey := acctest.RandString(8)
	attributeKey := acctest.RandString(8)
	attributeValue := acctest.RandString(8)
	attributeValue2 := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbItemAttributeDestroy(hashKey, rangeKey),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbItemAttributeConfigWithRangeKey(tableName, hashKey, rangeKey, attributeKey, attributeValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", "rangeKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue),
				),
			},
			{
				Config: testAccAWSDynamoDbItemAttributeConfigWithRangeKey(tableName, hashKey, rangeKey, attributeKey, attributeValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDynamoDbTableItemAttributeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table_name", tableName),
					resource.TestCheckResourceAttr(resourceName, "hash_key_value", "hashKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "range_key_value", "rangeKeyValue"),
					resource.TestCheckResourceAttr(resourceName, "attribute_key", attributeKey),
					resource.TestCheckResourceAttr(resourceName, "attribute_value", attributeValue2),
				),
			},
		},
	})
}

func testAccCheckAWSDynamoDbItemAttributeDestroy(hashKeyName, rangeKeyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_table_item_attribute" {
				continue
			}

			attrs := rs.Primary.Attributes

			result, err := conn.GetItem(&dynamodb.GetItemInput{
				ConsistentRead: aws.Bool(true),
				ExpressionAttributeNames: map[string]*string{
					"#key": aws.String(attrs["attribute_key"]),
				},
				Key:                  resourceAwsDynamoDbTableItemAttributeGetQueryKey(hashKeyName, attrs["hash_key_value"], rangeKeyName, attrs["range_key_value"]),
				ProjectionExpression: aws.String("#key"),
				TableName:            aws.String(attrs["table_name"]),
			})
			if err != nil {
				if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
					return nil
				}
				return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
			}
			if result.Item == nil {
				return nil
			}

			return fmt.Errorf("DynamoDB table item attribute %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAWSDynamoDbTableItemAttributeExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table item ID specified")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		attrs := rs.Primary.Attributes

		hashKeyName, rangeKeyName, err := resourceAwsDynamoDbTableItemAttributeGetKeysInfo(conn, attrs["table_name"])
		if err != nil {
			return err
		}

		result, err := conn.GetItem(&dynamodb.GetItemInput{
			ConsistentRead: aws.Bool(true),
			ExpressionAttributeNames: map[string]*string{
				"#key": aws.String(attrs["attribute_key"]),
			},
			Key:                  resourceAwsDynamoDbTableItemAttributeGetQueryKey(hashKeyName, attrs["hash_key_value"], rangeKeyName, attrs["range_key_value"]),
			ProjectionExpression: aws.String("#key"),
			TableName:            aws.String(attrs["table_name"]),
		})
		if err != nil {
			return fmt.Errorf("Problem getting table item '%s': %s", rs.Primary.ID, err)
		}

		expectedValue := attrs["attribute_value"]
		gottenValue := *result.Item[attrs["attribute_key"]].S
		if expectedValue != gottenValue {
			return fmt.Errorf("Got attribute value: %s, Expected: %s", gottenValue, expectedValue)
		}

		return nil
	}
}

func testAccAWSDynamoDbItemAttributeConfigBasic(tableName, hashKey, attributeKey, attributeValue string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name = "%[1]s"
  read_capacity = 10
  write_capacity = 10
  hash_key = "%[2]s"

  attribute {
    name = "%[2]s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key = "${aws_dynamodb_table.test.hash_key}"
  item = <<ITEM
{
	"%[2]s": {"S": "hashKeyValue"}
}
ITEM
}

resource "aws_dynamodb_table_item_attribute" "test" {
  table_name      = "${aws_dynamodb_table.test.name}"
  hash_key_value  = "hashKeyValue"
  attribute_key   = "%[3]s"
  attribute_value = "%[4]s"
  depends_on      = ["aws_dynamodb_table_item.test"]
}
`, tableName, hashKey, attributeKey, attributeValue)
}

func testAccAWSDynamoDbItemAttributeConfigWithRangeKey(tableName, hashKey, rangeKey, attributeKey, attributeValue string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name = "%[1]s"
  read_capacity = 10
  write_capacity = 10
  hash_key = "%[2]s"
  range_key = "%[3]s"

  attribute {
    name = "%[2]s"
    type = "S"
  }
  attribute {
    name = "%[3]s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = "${aws_dynamodb_table.test.name}"
  hash_key = "${aws_dynamodb_table.test.hash_key}"
  range_key = "${aws_dynamodb_table.test.range_key}"
  item = <<ITEM
{
	"%[2]s": {"S": "hashKeyValue"},
	"%[3]s": {"S": "rangeKeyValue"}
}
ITEM
}

resource "aws_dynamodb_table_item_attribute" "test" {
  table_name      = "${aws_dynamodb_table.test.name}"
  hash_key_value  = "hashKeyValue"
  range_key_value = "rangeKeyValue"
  attribute_key   = "%[4]s"
  attribute_value = "%[5]s"
  depends_on      = ["aws_dynamodb_table_item.test"]
}
`, tableName, hashKey, rangeKey, attributeKey, attributeValue)
}
