package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsDynamoDbTable_basic(t *testing.T) {
	tableName := fmt.Sprintf("testaccawsdynamodbtable-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsDynamoDbTableConfigBasic(tableName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "name", tableName),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "read_capacity", "20"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "write_capacity", "20"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "hash_key", "UserId"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "range_key", "GameTitle"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "attribute.#", "3"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "global_secondary_index.#", "1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "tags.%", "2"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "tags.Name", "dynamodb-table-1"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "tags.Environment", "test"),
					resource.TestCheckResourceAttr("data.aws_dynamodb_table.dynamodb_table_test", "server_side_encryption.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsDynamoDbTableConfigBasic(tableName string) string {
	return fmt.Sprintf(`resource "aws_dynamodb_table" "dynamodb_table_test" {
  name           = "%s"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "UserId"
  range_key      = "GameTitle"

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

  tags = {
    Name        = "dynamodb-table-1"
    Environment = "test"
  }
}

data "aws_dynamodb_table" "dynamodb_table_test" {
  name = "${aws_dynamodb_table.dynamodb_table_test.name}"
}`, tableName)
}
