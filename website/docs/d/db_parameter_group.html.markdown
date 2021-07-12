---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_parameter_group"
description: |-
  Provides details about an RDS Database Parameter Group.
---

# Data Source: aws_backup_vault

Use this data source to get information on an existing database parameter group.

## Example Usage

```hcl
data "aws_db_parameter_group" "example" {
  name = "name-of-parameter"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database parameter group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the DB parameter group.
* `family` - The name of the DB parameter group family that this DB parameter is compatible with.
* `description` - The customer specified description for the DB parameter group.
* `tags` - Metadata that you can assign to help organize the resources that you create.