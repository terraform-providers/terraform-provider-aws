---
layout: "aws"
page_title: "AWS: dynamodb_table_item"
sidebar_current: "docs-aws-resource-dynamodb-table-item"
description: |-
  Provides a DynamoDB table item resource
---

# aws_dynamodb_table_item

Provides a DynamoDB table item resource

-> **Note:** This resource is not meant to be used for managing large amounts of data in your table, it is not designed to scale.
  You should perform **regular backups** of all data in the table, see [AWS docs for more](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/BackupRestore.html).

## Example Usage

```hcl
resource "aws_dynamodb_table_item" "example" {
  table_name = "${aws_dynamodb_table.example.name}"
  hash_key   = "${aws_dynamodb_table.example.hash_key}"

  item = <<ITEM
{
  "exampleHashKey": {"S": "something"},
  "one": {"N": "11111"},
  "two": {"N": "22222"},
  "three": {"N": "33333"},
  "four": {"N": "44444"}
}
ITEM
}

resource "aws_dynamodb_table" "example" {
  name           = "example-name"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "exampleHashKey"

  attribute {
    name = "exampleHashKey"
    type = "S"
  }
}
```

## Argument Reference

The following arguments are supported:

* `table_name` - (Required) The name of the table to contain the item.
* `hash_key` - (Required) Hash key to use for lookups and identification of the item
* `range_key` - (Optional) Range key to use for lookups and identification of the item. Required if there is range key defined in the table.
* `item` - (Required) JSON representation of a map of attribute name/value pairs, one for each attribute.
  Only the primary key attributes are required; you can optionally provide other attribute name-value pairs for the item.

## Attributes Reference

All of the arguments above are exported as attributes.

## Import

DynamoDB table items cannot be imported.
