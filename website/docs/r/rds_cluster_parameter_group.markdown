---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_cluster_parameter_group"
description: |-
  Provides an RDS DB cluster parameter group resource.
---

# Resource: aws_rds_cluster_parameter_group

Provides an RDS DB cluster parameter group resource. Documentation of the available parameters for various Aurora engines can be found at:

* [Aurora MySQL Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Reference.html)
* [Aurora PostgreSQL Parameters](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraPostgreSQL.Reference.html)

## Example Usage

```hcl
resource "aws_rds_cluster_parameter_group" "default" {
  name        = "rds-cluster-pg"
  family      = "aurora5.6"
  description = "RDS default cluster parameter group"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the DB cluster parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `family` - (Required) The family of the DB cluster parameter group.
* `description` - (Optional) The description of the DB cluster parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of DB parameters to apply. Note that parameters may differ from a family to an other. Full list of all parameters can be discovered via [`aws rds describe-db-cluster-parameters`](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-cluster-parameters.html) after initial creation of the group.
* `tags` - (Optional) A map of tags to assign to the resource.

Parameter blocks support the following:

* `name` - (Required) The name of the DB parameter.
* `value` - (Required) The value of the DB parameter.
* `apply_method` - (Optional) "immediate" (default), or "pending-reboot". Some
    engines can't apply some parameters without a reboot, and you will need to
    specify "pending-reboot" here.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The db cluster parameter group name.
* `arn` - The ARN of the db cluster parameter group.


## Import

RDS Cluster Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_rds_cluster_parameter_group.cluster_pg production-pg-1
```
