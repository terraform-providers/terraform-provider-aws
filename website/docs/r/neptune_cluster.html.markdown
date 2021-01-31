---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster"
description: |-
  Provides an Neptune Cluster Resource
---

# Resource: aws_neptune_cluster

Provides an Neptune Cluster Resource. A Cluster Resource defines attributes that are
applied to the entire cluster of Neptune Cluster Instances.

Changes to a Neptune Cluster can occur when you manually change a
parameter, such as `backup_retention_period`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

## Example Usage

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
```

~> **Note:** AWS Neptune does not support user name/password–based access control.
See the AWS [Docs](https://docs.aws.amazon.com/neptune/latest/userguide/limits.html) for more information.

## Argument Reference

The following arguments are supported:

* `apply_immediately` - (Optional) Specifies whether any cluster modifications are applied immediately, or during the next maintenance window. Default is `false`.
* `availability_zones` - (Optional) A list of EC2 Availability Zones that instances in the Neptune cluster can be created in.
* `backup_retention_period` - (Optional) The days to retain backups for. Default `1`
* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifier`.
* `enable_cloudwatch_logs_exports` - (Optional) A list of the log types this DB cluster is configured to export to Cloudwatch Logs. Currently only supports `audit`.
* `engine` - (Optional) The name of the database engine to be used for this Neptune cluster. Defaults to `neptune`.
* `engine_version` - (Optional) The database engine version.
* `final_snapshot_identifier` - (Optional) The name of your final Neptune snapshot when this Neptune cluster is deleted. If omitted, no final snapshot will be made.
* `iam_roles` - (Optional) A List of ARNs for the IAM roles to associate to the Neptune Cluster.
* `iam_database_authentication_enabled` - (Optional) Specifies whether or mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled.
* `kms_key_arn` - (Optional) The ARN for the KMS encryption key. When specifying `kms_key_arn`, `storage_encrypted` needs to be set to true.
* `neptune_subnet_group_name` - (Optional) A Neptune subnet group to associate with this Neptune instance.
* `neptune_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter. Time in UTC. Default: A 30-minute window selected at random from an 8-hour block of time per region. e.g. 04:00-09:00
* `preferred_maintenance_window` - (Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g. wed:04:00-wed:04:30
* `port` - (Optional) The port on which the Neptune accepts connections. Default is `8182`.
* `replication_source_identifier` - (Optional) ARN of a source Neptune cluster or Neptune instance if this Neptune cluster is to be created as a Read Replica.
* `skip_final_snapshot` - (Optional) Determines whether a final Neptune snapshot is created before the Neptune cluster is deleted. If true is specified, no Neptune snapshot is created. If false is specified, a Neptune snapshot is created before the Neptune cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a Neptune cluster snapshot, or the ARN when specifying a Neptune snapshot.
* `storage_encrypted` - (Optional) Specifies whether the Neptune cluster is encrypted. The default is `false` if not specified.
* `tags` - (Optional) A map of tags to assign to the Neptune cluster.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate with the Cluster
* `deletion_protection` - (Optional) A value that indicates whether the DB cluster has deletion protection enabled.The database can't be deleted when deletion protection is enabled. By default, deletion protection is disabled.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Neptune Cluster Amazon Resource Name (ARN)
* `cluster_resource_id` - The Neptune Cluster Resource ID
* `cluster_members` – List of Neptune Instances that are a part of this cluster
* `endpoint` - The DNS address of the Neptune instance
* `hosted_zone_id` - The Route53 Hosted Zone ID of the endpoint
* `id` - The Neptune Cluster Identifier
* `reader_endpoint` - A read-only endpoint for the Neptune cluster, automatically load-balanced across replicas
* `status` - The Neptune instance status

## Timeouts

`aws_neptune_cluster` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `120 minutes`) Used for Cluster creation
- `update` - (Default `120 minutes`) Used for Cluster modifications
- `delete` - (Default `120 minutes`) Used for destroying cluster. This includes any cleanup task during the destroying process.

## Import

`aws_neptune_cluster` can be imported by using the cluster identifier, e.g.

```
$ terraform import aws_neptune_cluster.example my-cluster
```
