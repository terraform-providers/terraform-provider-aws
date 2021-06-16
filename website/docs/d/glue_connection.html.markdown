---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_connection"
description: |-
  Get information on an AWS Glue Connection
---

# Data Source: aws_glue_connection

This data source can be used to fetch information about a specific Glue Connection.

## Example Usage

```terraform
data "aws_glue_connection" "example" {
  id = "123456789123:connection"
}
```

## Argument Reference

* `id` - (Required) A concatenation of the catalog ID and connection name. For example, if your account ID is
`123456789123` and the connection name is `conn` then the ID is `123456789123:conn`.

## Attributes Reference

* `arn` - The ARN of the Glue Connection.
* `catalog_id` - The catalog ID of the Glue Connection.
* `connection_type` - The type of Glue Connection.
* `description` – Description of the connection.
* `match_criteria` – A list of criteria that can be used in selecting this connection.
* `name` - The name of the Glue Connection.
* `physical_connection_requirements` - A map of physical connection requirements, such as VPC and SecurityGroup.
