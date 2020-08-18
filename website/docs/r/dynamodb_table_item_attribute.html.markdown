---
layout: "aws"
page_title: "AWS: dynamodb_table_item_attribute"
sidebar_current: "docs-aws-resource-dynamodb-table-item-attribute"
description: |-
  Provides a DynamoDB table item attribute resource
---

# aws_dynamodb_table_item_attribute

Provides a DynamoDB table item attribute resource. Only supports strings.

-> **Note:** This resource is not meant to be used for managing large amounts of data in your table, it is not designed to scale.
  You should perform **regular backups** of all data in the table, see [AWS docs for more](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/BackupRestore.html).

## Example Usage

```hcl
resource "aws_dynamodb_table_item_attribute" "example" {
  table_name      = "${aws_dynamodb_table.example.name}"
  hash_key_value  = "something"
  attribute_key   = "five"
  attribute_value = "55555"
}

resource "aws_dynamodb_table_item" "example" {
  table_name = "${aws_dynamodb_table.example.name}"
  hash_key = "${aws_dynamodb_table.example.hash_key}"
  item = <<ITEM
{
  "exampleHashKey": {"S": "something"},
  "one": {"S": "11111"},
  "two": {"S": "22222"},
  "three": {"S": "33333"},
  "four": {"S": "44444"}
}
ITEM
}

resource "aws_dynamodb_table" "example" {
  name = "example-name"
  read_capacity = 10
  write_capacity = 10
  hash_key = "exampleHashKey"

  attribute {
    name = "exampleHashKey"
    type = "S"
  }
}
```

## Argument Reference

The following arguments are supported:

* `table_name` - (Required) The name of the table that contains the item (row) to modify
* `hash_key_value` - (Required) The value of the item's hash key
* `range_key_value` - (Optional) The value of the item's range key (If the table defines one).
* `attribute_key` - (Required) The key of the new attribute that will be added to the item.
* `attribute_value` - (Required) The value of the new attribute that will be added to the item.

## Attributes Reference

All of the arguments above are exported as attributes.

## Import

DynamoDB table item attributes can be imported using the table name, hash key value, range key value and attribute key, e.g.

Without a range key:
```
$ terraform import aws_dynamodb_table_item_attribute.example table_name:hash_key_value::attribute_key
```

With a range key:
```
$ terraform import aws_dynamodb_table_item_attribute.example table_name:hash_key_value:range_key_value:attribute_key
```
