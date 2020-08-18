---
layout: "aws"
page_title: "AWS: dynamodb_table_gsi"
sidebar_current: "docs-aws-resource-dynamodb-table-gsi"
description: |-
  Provides a DynamoDB Global Secondary Index resource
---

# aws_dynamodb_table_gsi

Provides a DynamoDB Global Secondary Index resource

~> **Note:** It is recommended to use `lifecycle` [`ignore_changes`](/docs/configuration/resources.html#ignore_changes) for `read_capacity` and/or `write_capacity` if there's [autoscaling policy](/docs/providers/aws/r/appautoscaling_policy.html) attached to the index.

## Example Usage

The following dynamodb table description models the table and GSI shown
in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)

```hcl
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = "GameScores"
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

  ttl {
    attribute_name = "TimeToExist"
    enabled        = false
  }

  tags {
    Name        = "dynamodb-table-1"
    Environment = "production"
  }
}

resource "aws_dynamodb_table_gsi" "basic-dynamodb-table-gsi" {
  name               = "GameTitleIndex"
  hash_key           = "GameTitle"
  range_key          = "TopScore"
  write_capacity     = 10
  read_capacity      = 10
  projection_type    = "INCLUDE"
  non_key_attributes = ["UserId"]

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }
}
```

Notes: `attribute` can be lists

```
  attribute = [{
    name = "UserId"
    type = "S"
  }, {
    name = "GameTitle"
    type = "S"
  }, {
    name = "TopScore"
    type = "N"
  }]
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the GSI, this needs to be unique
  within the table definition.
* `hash_key` - (Required, Forces new resource) The attribute to use as the hash (partition) key for the GSI. Must also be defined as an `attribute`, see below.
* `range_key` - (Optional, Forces new resource) The attribute to use as the range (sort) key for the GSI. Must also be defined as an `attribute`, see below.
* `write_capacity` - (Required) The number of write units for this GSI
* `read_capacity` - (Required) The number of read units for this GSI
* `attribute` - (Required) List of nested attribute definitions. Only required for `hash_key` and `range_key` attributes. Each attribute has two properties:
  * `name` - (Required) The name of the attribute
  * `type` - (Required) Attribute type, which must be a scalar type: `S`, `N`, or `B` for (S)tring, (N)umber or (B)inary data
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY`
   where `ALL` projects every attribute into the index, `KEYS_ONLY`
    projects just the hash and range key into the index, and `INCLUDE`
    projects only the keys specified in the _non_key_attributes_
    parameter.
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a
  projection type; a list of attributes to project into the index. These
  do not need to be defined as attributes on the table.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when creating the GSI
* `update` - (Defaults to 5 mins) Used when updating the GSI. Note that a GSI update only consists of provisioned capacity updates
* `delete` - (Defaults to 10 mins) Used when deleting the GSI

### A note about attributes

Only define attributes on the table object that are going to be used as:

* GSI hash key or range key

The DynamoDB API expects attribute structure (name and type) to be
passed along when creating or updating GSI/LSIs or creating the initial
table. You should not add any attributes that belong to the table or
other indexes. If you add attributes here that are not used in these
scenarios it can cause an infinite loop in planning.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the index
