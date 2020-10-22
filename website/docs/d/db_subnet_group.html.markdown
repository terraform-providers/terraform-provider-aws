---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_subnet_group"
description: |-
  Get information on an RDS Database Subnet Group.
---

# Data Source: aws_db_subnet_group

Use this data source to get information about an RDS subnet group.

## Example Usage

```hcl
data "aws_db_subnet_group" "database" {
  name = "my-test-database-subnet-group"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the RDS database subnet group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for the DB subnet group.
* `description` - Provides the description of the DB subnet group.
* `status` - Provides the status of the DB subnet group.
* `subnet_ids` - Contains a list of subnet identifiers.
* `vpc_id` - Provides the VPC ID of the subnet group.
