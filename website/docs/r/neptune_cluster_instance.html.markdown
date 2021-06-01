---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_instance"
description: |-
  Provides an Neptune Cluster Resource Instance
---

# Resource: aws_neptune_cluster_instance

A Cluster Instance Resource defines attributes that are specific to a single instance in a Neptune Cluster.

You can simply add neptune instances and Neptune manages the replication. You can use the [count][1]
meta-parameter to make multiple instances and join them all to the same Neptune Cluster, or you may specify different Cluster Instance resources with various `instance_class` sizes.


## Example Usage

The following example will create a neptune cluster with two neptune instances(one writer and one reader).

```hcl
resource "aws_neptune_cluster" "default" {
  cluster_identifier                  = "neptune-cluster-demo"
  engine                              = "neptune"
  backup_retention_period             = 5
  preferred_backup_window             = "07:00-09:00"
  skip_final_snapshot                 = true
  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_neptune_cluster_instance" "example" {
  count              = 2
  cluster_identifier = aws_neptune_cluster.default.id
  engine             = "neptune"
  instance_class     = "db.r4.large"
  apply_immediately  = true
}
```

## Argument Reference

The following arguments are supported:

* `apply_immediately` - (Optional) Specifies whether any instance modifications
  are applied immediately, or during the next maintenance window. Default is`false`.
* `auto_minor_version_upgrade` - (Optional) Indicates that minor engine upgrades will be applied automatically to the instance during the maintenance window. Default is `true`.
* `availability_zone` - (Optional) The EC2 Availability Zone that the neptune instance is created in.
* `cluster_identifier` - (Required) The identifier of the [`aws_neptune_cluster`](/docs/providers/aws/r/neptune_cluster.html) in which to launch this instance.
* `engine` - (Optional) The name of the database engine to be used for the neptune instance. Defaults to `neptune`. Valid Values: `neptune`.
* `engine_version` - (Optional) The neptune engine version.
* `identifier` - (Optional, Forces new resource) The identifier for the neptune instance, if omitted, Terraform will assign a random, unique identifier.
* `identifier_prefix` - (Optional, Forces new resource) Creates a unique identifier beginning with the specified prefix. Conflicts with `identifier`.
* `instance_class` - (Required) The instance class to use.
* `neptune_subnet_group_name` - (Required if `publicly_accessible = false`, Optional otherwise) A subnet group to associate with this neptune instance. **NOTE:** This must match the `neptune_subnet_group_name` of the attached [`aws_neptune_cluster`](/docs/providers/aws/r/neptune_cluster.html).
* `neptune_parameter_group_name` - (Optional) The name of the neptune parameter group to associate with this instance.
* `port` - (Optional) The port on which the DB accepts connections. Defaults to `8182`.
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled. Eg: "04:00-09:00"
* `preferred_maintenance_window` - (Optional) The window to perform maintenance in.
  Syntax: "ddd:hh24:mi-ddd:hh24:mi". Eg: "Mon:00:00-Mon:03:00".
* `promotion_tier` - (Optional) Default 0. Failover Priority setting on instance level. The reader who has lower tier has higher priority to get promoter to writer.
* `publicly_accessible` - (Optional) Bool to control if instance is publicly accessible. Default is `false`.
* `tags` - (Optional) A map of tags to assign to the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `address` - The hostname of the instance. See also `endpoint` and `port`.
* `arn` - Amazon Resource Name (ARN) of neptune instance
* `dbi_resource_id` - The region-unique, immutable identifier for the neptune instance.
* `endpoint` - The connection endpoint in `address:port` format.
* `id` - The Instance identifier
* `kms_key_arn` - The ARN for the KMS encryption key if one is set to the neptune cluster.
* `storage_encrypted` - Specifies whether the neptune cluster is encrypted.
* `writer` – Boolean indicating if this instance is writable. `False` indicates this instance is a read replica.

[1]: https://www.terraform.io/docs/configuration/meta-arguments/count.html

## Timeouts

`aws_neptune_cluster_instance` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `90 minutes`) How long to wait for creating instances to become available.
- `update` - (Default `90 minutes`) How long to wait for updating instances to complete updates.
- `delete` - (Default `90 minutes`) How long to wait for deleting instances to become fully deleted.

## Import

`aws_neptune_cluster_instance` can be imported by using the instance identifier, e.g.

```
$ terraform import aws_neptune_cluster_instance.example my-instance
```
