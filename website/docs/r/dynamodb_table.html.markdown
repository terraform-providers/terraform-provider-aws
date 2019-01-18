---
layout: "aws"
page_title: "AWS: dynamodb_table"
sidebar_current: "docs-aws-resource-dynamodb-table"
description: |-
  Provides a DynamoDB table resource
---

# aws_dynamodb_table

Provides a DynamoDB table resource

~> **Note:** It is recommended to use `lifecycle` [`ignore_changes`](/docs/configuration/resources.html#ignore_changes) for `read_capacity` and/or `write_capacity` if there's [autoscaling policy](/docs/providers/aws/r/appautoscaling_policy.html) attached to the table.

## Example Usage

The following dynamodb table description models the table and GSI shown
in the [AWS SDK example documentation](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/GSI.html)

```hcl
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name           = "GameScores"
  billing_mode   = "PROVISIONED"
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

  ttl {
    attribute_name = "TimeToExist"
    enabled        = false
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
    Environment = "production"
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

* `name` - (Required) The name of the table, this needs to be unique
  within a region.
* `billing_mode` - (Optional) Controls how you are charged for read and write throughput and how you manage capacity. The valid values are `PROVISIONED` and `PAY_PER_REQUEST`. Defaults to `PROVISIONED`.
* `hash_key` - (Required, Forces new resource) The attribute to use as the hash (partition) key. Must also be defined as an `attribute`, see below.
* `range_key` - (Optional, Forces new resource) The attribute to use as the range (sort) key. Must also be defined as an `attribute`, see below.
* `write_capacity` - (Optional) The number of write units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.
* `read_capacity` - (Optional) The number of read units for this table. If the `billing_mode` is `PROVISIONED`, this field is required.
* `attribute` - (Required) List of nested attribute definitions. Only required for `hash_key` and `range_key` attributes. Each attribute has two properties:
  * `name` - (Required) The name of the attribute
  * `type` - (Required) Attribute type, which must be a scalar type: `S`, `N`, or `B` for (S)tring, (N)umber or (B)inary data
* `ttl` - (Optional) Defines ttl, has two properties, and can only be specified once:
  * `enabled` - (Required) Indicates whether ttl is enabled (true) or disabled (false).
  * `attribute_name` - (Required) The name of the table attribute to store the TTL timestamp in.
* `local_secondary_index` - (Optional, Forces new resource) Describe an LSI on the table;
  these can only be allocated *at creation* so you cannot change this
definition after you have created the resource.
* `global_secondary_index` - (Optional) Describe a GSO for the table;
  subject to the normal limits on the number of GSIs, projected
attributes, etc.
* `stream_enabled` - (Optional) Indicates whether Streams are to be enabled (true) or disabled (false).
* `stream_view_type` - (Optional) When an item in the table is modified, StreamViewType determines what information is written to the table's stream. Valid values are `KEYS_ONLY`, `NEW_IMAGE`, `OLD_IMAGE`, `NEW_AND_OLD_IMAGES`.
* `server_side_encryption` - (Optional) Encrypt at rest options.
* `tags` - (Optional) A map of tags to populate on the created table.
* `point_in_time_recovery` - (Optional) Point-in-time recovery options.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when creating the table
* `update` - (Defaults to 10 mins) Used when updating the table
* `delete` - (Defaults to 10 mins) Used when deleting the table

### Nested fields

#### `local_secondary_index`

* `name` - (Required) The name of the index
* `range_key` - (Required) The name of the range key; must be defined
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY`
   where `ALL` projects every attribute into the index, `KEYS_ONLY`
    projects just the hash and range key into the index, and `INCLUDE`
    projects only the keys specified in the _non_key_attributes_
    parameter.
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a
  projection type; a list of attributes to project into the index. These
  do not need to be defined as attributes on the table.

#### `global_secondary_index`

* `name` - (Required) The name of the index
* `write_capacity` - (Optional) The number of write units for this index. Must be set if billing_mode is set to PROVISIONED.
* `read_capacity` - (Optional) The number of read units for this index. Must be set if billing_mode is set to PROVISIONED.
* `hash_key` - (Required) The name of the hash key in the index; must be
  defined as an attribute in the resource.
* `range_key` - (Optional) The name of the range key; must be defined
* `projection_type` - (Required) One of `ALL`, `INCLUDE` or `KEYS_ONLY`
   where `ALL` projects every attribute into the index, `KEYS_ONLY`
    projects just the hash and range key into the index, and `INCLUDE`
    projects only the keys specified in the _non_key_attributes_
    parameter.
* `non_key_attributes` - (Optional) Only required with `INCLUDE` as a
  projection type; a list of attributes to project into the index. These
  do not need to be defined as attributes on the table.

#### `server_side_encryption`

* `enabled` - (Required) Whether to enable encryption at rest. If the `server_side_encryption` block is not provided then this defaults to `false`.

#### `point_in_time_recovery`

* `enabled` - (Required) Whether to enable point-in-time recovery - note that it can take up to 10 minutes to enable for new tables. If the `point_in_time_recovery` block is not provided then this defaults to `false`.

### A note about attributes

Only define attributes on the table object that are going to be used as:

* Table hash key or range key
* LSI or GSI hash key or range key

The DynamoDB API expects attribute structure (name and type) to be
passed along when creating or updating GSI/LSIs or creating the initial
table. In these cases it expects the Hash / Range keys to be provided;
because these get re-used in numerous places (i.e the table's range key
could be a part of one or more GSIs), they are stored on the table
object to prevent duplication and increase consistency. If you add
attributes here that are not used in these scenarios it can cause an
infinite loop in planning.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The arn of the table
* `id` - The name of the table
* `stream_arn` - The ARN of the Table Stream. Only available when `stream_enabled = true`
* `stream_label` - A timestamp, in ISO 8601 format, for this stream. Note that this timestamp is not
  a unique identifier for the stream on its own. However, the combination of AWS customer ID,
  table name and this field is guaranteed to be unique.
  It can be used for creating CloudWatch Alarms. Only available when `stream_enabled = true`

## Import

DynamoDB tables can be imported using the `name`, e.g.

```
$ terraform import aws_dynamodb_table.basic-dynamodb-table GameScores
```
