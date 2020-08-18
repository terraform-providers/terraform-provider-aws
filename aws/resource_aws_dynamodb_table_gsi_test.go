package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDynamoDbTableGsi_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := acctest.RandomWithPrefix("TerraformTestTable")
	gsiName := acctest.RandomWithPrefix("TerraformTestGsi")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.table", "name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "hash_key", "TestGsiHashKey"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "projection_type", "KEYS_ONLY"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_updateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := acctest.RandomWithPrefix("TerraformTestTable")
	gsiName := acctest.RandomWithPrefix("TerraformTestGsi")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbGsiConfigBasicMoreCapacity(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),

					// TODO: check gsi was not re-created; not sure how to do this atm
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "5"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "5"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_updateOtherAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := acctest.RandomWithPrefix("TerraformTestTable")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiInitialOtherAttributes(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),

					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", "gsi1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "hash_key", "att1"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi", "range_key"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi", "non_key_attributes"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),

					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "name", "gsi2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "hash_key", "att2"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi2", "range_key"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi2", "non_key_attributes"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbGsiConfigUpdatedOtherAttributes(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),

					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", "gsi-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "hash_key", "att1"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi", "range_key"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "non_key_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),

					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "name", "gsi2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "range_key", "att2"),
					resource.TestCheckNoResourceAttr("aws_dynamodb_table_gsi.gsi2", "non_key_attributes"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi2", "write_capacity", "1"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTableGsi_delete(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	tableName := acctest.RandomWithPrefix("TerraformTestTable")
	gsiName := acctest.RandomWithPrefix("TerraformTestGsi")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.table", "name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "name", gsiName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table_gsi.gsi", "write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbGsiConfigBasicDelete(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.table", &conf),
					func(*terraform.State) error {
						if len(conf.Table.GlobalSecondaryIndexes) > 0 {
							return fmt.Errorf("expected to find no global secondary indexes, instead found: %v", conf.Table.GlobalSecondaryIndexes)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccAWSDynamoDbGsiConfigBasic(tableName, gsiName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi" {
  table_name = "${aws_dynamodb_table.table.id}"
  name = "%s"
  hash_key = "TestGsiHashKey"
  write_capacity = 1
  read_capacity = 1
  projection_type = "KEYS_ONLY"

  attribute {
    name = "TestGsiHashKey"
    type = "S"
  }
}
`, tableName, gsiName)
}

func testAccAWSDynamoDbGsiConfigBasicMoreCapacity(tableName, gsiName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi" {
  table_name = "${aws_dynamodb_table.table.id}"
  name = "%s"
  hash_key = "TestGsiHashKey"
  write_capacity = 5
  read_capacity = 5
  projection_type = "KEYS_ONLY"

  attribute {
    name = "TestGsiHashKey"
    type = "S"
  }
}
`, tableName, gsiName)
}

func testAccAWSDynamoDbGsiConfigBasicDelete(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, tableName)
}

func testAccAWSDynamoDbGsiInitialOtherAttributes(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name           = "%s"
  read_capacity  = "1"
  write_capacity = "1"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi" {
  table_name = "${aws_dynamodb_table.table.id}"
  name            = "gsi1-index"
  hash_key        = "att1"
  write_capacity  = "1"
  read_capacity   = "1"
  projection_type = "ALL"

  attribute {
    name = "att1"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi2" {
  table_name = "${aws_dynamodb_table.table.id}"
  name            = "gsi2-index"
  hash_key        = "att2"
  write_capacity  = "1"
  read_capacity   = "1"
  projection_type = "ALL"

  attribute {
    name = "att2"
    type = "S"
  }
}
`, tableName)
}

func testAccAWSDynamoDbGsiConfigUpdatedOtherAttributes(tableName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "table" {
  name           = "%s"
  read_capacity  = "1"
  write_capacity = "1"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi" {
  table_name         = "${aws_dynamodb_table.table.id}"
  name               = "gsi1-index"
  hash_key           = "att1"
  write_capacity     = "1"
  read_capacity      = "1"
  projection_type    = "INCLUDE"
  non_key_attributes = ["RandomAttribute"]
  
  attribute {
    name = "att1"
    type = "S"
  }
}

resource "aws_dynamodb_table_gsi" "gsi2" {
  table_name      = "${aws_dynamodb_table.table.id}"
  name            = "gsi2-index"
  hash_key        = "att3"
  range_key       = "att2"
  write_capacity  = "1"
  read_capacity   = "1"
  projection_type = "ALL"

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "N"
  }
}
`, tableName)
}
