---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_global_table"
description: |-
  Manages DynamoDB Global Tables V1 (version 2017.11.29)
---

# Resource: aws_dynamodb_global_table

Manages [DynamoDB Global Tables V1 (version 2017.11.29)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V1.html). These are layered on top of existing DynamoDB Tables.

~> **NOTE:** To instead manage [DynamoDB Global Tables V2 (version 2019.11.21)](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables.V2.html), use the [`aws_dynamodb_table` resource](/docs/providers/aws/r/dynamodb_table.html) `replica` configuration block.

~> Note: There are many restrictions before you can properly create DynamoDB Global Tables in multiple regions. See the [AWS DynamoDB Global Table Requirements](http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/globaltables_reqs_bestpractices.html) for more information.

## Example Usage

```hcl
provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

provider "aws" {
  alias  = "us-west-2"
  region = "us-west-2"
}

resource "aws_dynamodb_table" "us-east-1" {
  provider = aws.us-east-1

  hash_key         = "myAttribute"
  name             = "myTable"
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
  provider = aws.us-west-2

  hash_key         = "myAttribute"
  name             = "myTable"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"
  read_capacity    = 1
  write_capacity   = 1

  attribute {
    name = "myAttribute"
    type = "S"
  }
}

resource "aws_dynamodb_global_table" "myTable" {
  depends_on = [
    aws_dynamodb_table.us-east-1,
    aws_dynamodb_table.us-west-2,
  ]
  provider = aws.us-east-1

  name = "myTable"

  replica {
    region_name = "us-east-1"
  }

  replica {
    region_name = "us-west-2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the global table. Must match underlying DynamoDB Table names in all regions.
* `replica` - (Required) Underlying DynamoDB Table. At least 1 replica must be defined. See below.

### Nested Fields

#### `replica`

* `region_name` - (Required) AWS region name of replica DynamoDB Table. e.g. `us-east-1`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the DynamoDB Global Table
* `arn` - The ARN of the DynamoDB Global Table

## Import

DynamoDB Global Tables can be imported using the global table name, e.g.

```
$ terraform import aws_dynamodb_global_table.MyTable MyTable
```
