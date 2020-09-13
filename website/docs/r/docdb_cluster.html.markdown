---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb"
description: |-
  Manages a DocDB Aurora Cluster
---

# Resource: aws_docdb_cluster

Manages a DocDB Cluster.

Changes to a DocDB Cluster can occur when you manually change a
parameter, such as `port`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

~> **Note:** using `apply_immediately` can result in a brief downtime as the server reboots.
~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
resource "aws_docdb_cluster" "docdb" {
  cluster_identifier      = "my-docdb-cluster"
  engine                  = "docdb"
  master_username         = "foo"
  master_password         = "mustbeeightchars"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  skip_final_snapshot     = true
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/docdb/create-db-cluster.html).

The following arguments are supported:

* `apply_immediately` - (Optional) Specifies whether any cluster modifications
     are applied immediately, or during the next maintenance window. Default is
     `false`.
* `availability_zones` - (Optional) A list of EC2 Availability Zones that
  instances in the DB cluster can be created in.
* `backup_retention_period` - (Optional) The days to retain backups for. Default `1`
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifier`.
* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `db_subnet_group_name` - (Optional) A DB subnet group to associate with this DB instance.
* `db_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `deletion_protection` - (Optional) A value that indicates whether the DB cluster has deletion protection enabled. The database can't be deleted when deletion protection is enabled. By default, deletion protection is disabled.
* `enabled_cloudwatch_logs_exports` - (Optional) List of log types to export to cloudwatch. If omitted, no logs will be exported.
   The following log types are supported: `audit`, `profiler`.
* `engine_version` - (Optional) The database engine version. Updating this argument results in an outage.
* `engine` - (Optional) The name of the database engine to be used for this DB cluster. Defaults to `docdb`. Valid Values: `docdb`
* `final_snapshot_identifier` - (Optional) The name of your final DB snapshot
    when this DB cluster is deleted. If omitted, no final snapshot will be
    made.
* `kms_key_id` - (Optional) The ARN for the KMS encryption key. When specifying `kms_key_id`, `storage_encrypted` needs to be set to true.
* `master_password` - (Required unless a `snapshot_identifier` is provided) Password for the master DB user. Note that this may
    show up in logs, and it will be stored in the state file. Please refer to the DocDB Naming Constraints.
* `master_username` - (Required unless a `snapshot_identifier` is provided) Username for the master DB user.
* `port` - (Optional) The port on which the DB accepts connections
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC
Default: A 30-minute window selected at random from an 8-hour block of time per region. e.g. 04:00-09:00
* `skip_final_snapshot` - (Optional) Determines whether a final DB snapshot is created before the DB cluster is deleted. If true is specified, no DB snapshot is created. If false is specified, a DB snapshot is created before the DB cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a DB cluster snapshot, or the ARN when specifying a DB snapshot.
* `storage_encrypted` - (Optional) Specifies whether the DB cluster is encrypted. The default is `false`.
* `tags` - (Optional) A map of tags to assign to the DB cluster.
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate
  with the Cluster

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of cluster
* `cluster_members` – List of DocDB Instances that are a part of this cluster
* `cluster_resource_id` - The DocDB Cluster Resource ID
* `endpoint` - The DNS address of the DocDB instance
* `hosted_zone_id` - The Route53 Hosted Zone ID of the endpoint
* `id` - The DocDB Cluster Identifier
* `maintenance_window` - The instance maintenance window
* `reader_endpoint` - A read-only endpoint for the DocDB cluster, automatically load-balanced across replicas

## Timeouts

`aws_docdb_cluster` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `120 minutes`) Used for Cluster creation
- `update` - (Default `120 minutes`) Used for Cluster modifications
- `delete` - (Default `120 minutes`) Used for destroying cluster. This includes
any cleanup task during the destroying process.

## Import

DocDB Clusters can be imported using the `cluster_identifier`, e.g.

```
$ terraform import aws_docdb_cluster.docdb_cluster docdb-prod-cluster
```
