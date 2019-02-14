---
layout: "aws"
page_title: "AWS: aws_dynamodb_table"
sidebar_current: "docs-aws-datasource-dynamodb-table"
description: |-
  Provides a DynamoDB table data source.
---

# Data Source: aws_dynamodb_table

Provides information about a DynamoDB table.

## Example Usage

```hcl
data "aws_dynamodb_table" "tableName" {
  name = "tableName"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the DynamoDB table.

## Attributes Reference

See the [DynamoDB Table Resource](/docs/providers/aws/r/dynamodb_table.html) for details on the
returned attributes - they are identical.
