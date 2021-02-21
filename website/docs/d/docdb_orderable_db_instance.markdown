---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_orderable_db_instance"
description: |-
  Information about DocumentDB orderable DB instances.
---

# Data Source: aws_docdb_orderable_db_instance

Information about DocumentDB orderable DB instances.

## Example Usage

```hcl
data "aws_docdb_orderable_db_instance" "test" {
  engine         = "docdb"
  engine_version = "3.6.0"
  license_model  = "na"

  preferred_instance_classes = ["db.r5.large", "db.r4.large", "db.t3.medium"]
}
```

## Argument Reference

The following arguments are supported:

* `engine` - (Optional) DB engine. Default: `docdb`
* `engine_version` - (Optional) Version of the DB engine.
* `instance_class` - (Optional) DB instance class. Examples of classes are `db.r5.12xlarge`, `db.r5.24xlarge`, `db.r5.2xlarge`, `db.r5.4xlarge`, `db.r5.large`, `db.r5.xlarge`, and `db.t3.medium`. (Conflicts with `preferred_instance_classes`.)
* `license_model` - (Optional) License model. Default: `na`
* `preferred_instance_classes` - (Optional) Ordered list of preferred DocumentDB DB instance classes. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. (Conflicts with `instance_class`.)
* `vpc` - (Optional) Enable to show only VPC.

## Attribute Reference

In addition to all arguments above, the following attribute is exported:

* `availability_zones` - Availability zones where the instance is available.
