---
layout: "aws"
page_title: "AWS: aws_rds_cluster"
sidebar_current: "docs-aws-resource-rds-cluster"
description: |-
  Manages a RDS Aurora Cluster
---

# aws_rds_cluster

Manages a [RDS Aurora Cluster][2]. To manage cluster instances that inherit configuration from the cluster (when not running the cluster in `serverless` engine mode), see the [`aws_rds_cluster_instance` resource](/docs/providers/aws/r/rds_cluster_instance.html). To manage non-Aurora databases (e.g. MySQL, PostgreSQL, SQL Server, etc.), see the [`aws_db_instance` resource](/docs/providers/aws/r/db_instance.html).

For information on the difference between the available Aurora MySQL engines
see [Comparison between Aurora MySQL 1 and Aurora MySQL 2](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Updates.20180206.html)
in the Amazon RDS User Guide.

Changes to a RDS Cluster can occur when you manually change a
parameter, such as `port`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

~> **Note:** using `apply_immediately` can result in a
brief downtime as the server reboots. See the AWS Docs on [RDS Maintenance][4]
for more information.

~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

### Aurora MySQL 2.x (MySQL 5.7)

```hcl
resource "aws_rds_cluster" "default" {
  cluster_identifier      = "aurora-cluster-demo"
  engine                  = "aurora-mysql"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "bar"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

### Aurora MySQL 1.x (MySQL 5.6)

```hcl
resource "aws_rds_cluster" "default" {
  cluster_identifier      = "aurora-cluster-demo"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "bar"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

### Aurora with PostgreSQL engine

```hcl
resource "aws_rds_cluster" "postgresql" {
  cluster_identifier      = "aurora-cluster-demo"
  engine                  = "aurora-postgresql"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "bar"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster.html).

The following arguments are supported:

* `cluster_identifier` - (Optional, Forces new resources) The cluster identifier. If omitted, Terraform will assign a random, unique identifier.
* `cluster_identifier_prefix` - (Optional, Forces new resource) Creates a unique cluster identifier beginning with the specified prefix. Conflicts with `cluster_identifer`.
* `database_name` - (Optional) Name for an automatically created database on cluster creation. There are different naming restrictions per database engine: [RDS Naming Constraints][5]
* `deletion_protection` - (Optional) If the DB instance should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `master_password` - (Required unless a `snapshot_identifier` is provided) Password for the master DB user. Note that this may
    show up in logs, and it will be stored in the state file. Please refer to the [RDS Naming Constraints][5]
* `master_username` - (Required unless a `snapshot_identifier` is provided) Username for the master DB user. Please refer to the [RDS Naming Constraints][5]
* `final_snapshot_identifier` - (Optional) The name of your final DB snapshot
    when this DB cluster is deleted. If omitted, no final snapshot will be
    made.
* `skip_final_snapshot` - (Optional) Determines whether a final DB snapshot is created before the DB cluster is deleted. If true is specified, no DB snapshot is created. If false is specified, a DB snapshot is created before the DB cluster is deleted, using the value from `final_snapshot_identifier`. Default is `false`.
* `availability_zones` - (Optional) A list of EC2 Availability Zones that
  instances in the DB cluster can be created in
* `backtrack_window` - (Optional) The target backtrack window, in seconds. Only available for `aurora` engine currently. To disable backtracking, set this value to `0`. Defaults to `0`. Must be between `0` and `259200` (72 hours)
* `backup_retention_period` - (Optional) The days to retain backups for. Default `1`
* `preferred_backup_window` - (Optional) The daily time range during which automated backups are created if automated backups are enabled using the BackupRetentionPeriod parameter.Time in UTC
Default: A 30-minute window selected at random from an 8-hour block of time per region. e.g. 04:00-09:00
* `preferred_maintenance_window` - (Optional) The weekly time range during which system maintenance can occur, in (UTC) e.g. wed:04:00-wed:04:30
* `port` - (Optional) The port on which the DB accepts connections
* `vpc_security_group_ids` - (Optional) List of VPC security groups to associate
  with the Cluster
* `snapshot_identifier` - (Optional) Specifies whether or not to create this cluster from a snapshot. You can use either the name or ARN when specifying a DB cluster snapshot, or the ARN when specifying a DB snapshot.
* `storage_encrypted` - (Optional) Specifies whether the DB cluster is encrypted. The default is `false` for `provisioned` `engine_mode` and `true` for `serverless` `engine_mode`.
* `replication_source_identifier` - (Optional) ARN of a source DB cluster or DB instance if this DB cluster is to be created as a Read Replica.
* `apply_immediately` - (Optional) Specifies whether any cluster modifications
     are applied immediately, or during the next maintenance window. Default is
     `false`. See [Amazon RDS Documentation for more information.](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.DBInstance.Modifying.html)
* `db_subnet_group_name` - (Optional) A DB subnet group to associate with this DB instance. **NOTE:** This must match the `db_subnet_group_name` specified on every [`aws_rds_cluster_instance`](/docs/providers/aws/r/rds_cluster_instance.html) in the cluster.
* `db_cluster_parameter_group_name` - (Optional) A cluster parameter group to associate with the cluster.
* `kms_key_id` - (Optional) The ARN for the KMS encryption key. When specifying `kms_key_id`, `storage_encrypted` needs to be set to true.
* `iam_roles` - (Optional) A List of ARNs for the IAM roles to associate to the RDS Cluster.
* `iam_database_authentication_enabled` - (Optional) Specifies whether or mappings of AWS Identity and Access Management (IAM) accounts to database accounts is enabled. Please see [AWS Documentation][6] for availability and limitations.
* `engine` - (Optional) The name of the database engine to be used for this DB cluster. Defaults to `aurora`. Valid Values: `aurora`, `aurora-mysql`, `aurora-postgresql`
* `engine_mode` - (Optional) The database engine mode. Valid values: `global`, `parallelquery`, `provisioned`, `serverless`. Defaults to: `provisioned`. See the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/aurora-serverless.html) for limitations when using `serverless`.
* `engine_version` - (Optional) The database engine version. Updating this argument results in an outage.
* `source_region` - (Optional) The source region for an encrypted replica DB cluster.
* `enabled_cloudwatch_logs_exports` - (Optional) List of log types to export to cloudwatch. If omitted, no logs will be exported.
   The following log types are supported: `audit`, `error`, `general`, `slowquery`.
* `scaling_configuration` - (Optional) Nested attribute with scaling properties. Only valid when `engine_mode` is set to `serverless`. More details below.
* `tags` - (Optional) A mapping of tags to assign to the DB cluster.

### S3 Import Options

Full details on the core parameters and impacts are in the API Docs: [RestoreDBClusterFromS3](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBClusterFromS3.html). Requires that the S3 bucket be in the same region as the RDS cluster you're trying to create. Sample:

~> **NOTE:** RDS Aurora Serverless does not support loading data from S3, so its not possible to directly use `engine_mode` set to `serverless` with `s3_import`.

```hcl
resource "aws_rds_cluster" "db" {
  engine = "aurora"

  s3_import {
    source_engine         = "mysql"
    source_engine_version = "5.6"
    bucket_name           = "mybucket"
    bucket_prefix         = "backups"
    ingestion_role        = "arn:aws:iam::1234567890:role/role-xtrabackup-rds-restore"
  }
}
```

* `bucket_name` - (Required) The bucket name where your backup is stored
* `bucket_prefix` - (Optional) Can be blank, but is the path to your backup
* `ingestion_role` - (Required) Role applied to load the data.
* `source_engine` - (Required) Source engine for the backup
* `source_engine_version` - (Required) Version of the source engine used to make the backup

This will not recreate the resource if the S3 object changes in some way. It's only used to initialize the database. This only works currently with the aurora engine. See AWS for currently supported engines and options. See [Aurora S3 Migration Docs](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Migrating.ExtMySQL.html#AuroraMySQL.Migrating.ExtMySQL.S3).

### scaling_configuration Argument Reference

~> **NOTE:** `scaling_configuration` configuration is only valid when `engine_mode` is set to `serverless`.

Example:

```hcl
resource "aws_rds_cluster" "example" {
  # ... other configuration ...

  engine_mode = "serverless"

  scaling_configuration {
    auto_pause               = true
    max_capacity             = 256
    min_capacity             = 2
    seconds_until_auto_pause = 300
  }
}
```

* `auto_pause` - (Optional) Whether to enable automatic pause. A DB cluster can be paused only when it's idle (it has no connections). If a DB cluster is paused for more than seven days, the DB cluster might be backed up with a snapshot. In this case, the DB cluster is restored when there is a request to connect to it. Defaults to `true`.
* `max_capacity` - (Optional) The maximum capacity. The maximum capacity must be greater than or equal to the minimum capacity. Valid capacity values are `2`, `4`, `8`, `16`, `32`, `64`, `128`, and `256`. Defaults to `16`.
* `min_capacity` - (Optional) The minimum capacity. The minimum capacity must be lesser than or equal to the maximum capacity. Valid capacity values are `2`, `4`, `8`, `16`, `32`, `64`, `128`, and `256`. Defaults to `2`.
* `seconds_until_auto_pause` - (Optional) The time, in seconds, before an Aurora DB cluster in serverless mode is paused. Valid values are `300` through `86400`. Defaults to `300`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of cluster
* `id` - The RDS Cluster Identifier
* `cluster_identifier` - The RDS Cluster Identifier
* `cluster_resource_id` - The RDS Cluster Resource ID
* `cluster_members` – List of RDS Instances that are a part of this cluster
* `allocated_storage` - The amount of allocated storage
* `availability_zones` - The availability zone of the instance
* `backup_retention_period` - The backup retention period
* `preferred_backup_window` - The daily time range during which the backups happen
* `preferred_maintenance_window` - The maintenance window
* `endpoint` - The DNS address of the RDS instance
* `reader_endpoint` - A read-only endpoint for the Aurora cluster, automatically
load-balanced across replicas
* `engine` - The database engine
* `engine_version` - The database engine version
* `maintenance_window` - The instance maintenance window
* `database_name` - The database name
* `port` - The database port
* `status` - The RDS instance status
* `master_username` - The master username for the database
* `storage_encrypted` - Specifies whether the DB cluster is encrypted
* `replication_source_identifier` - ARN of the source DB cluster or DB instance if this DB cluster is created as a Read Replica.
* `hosted_zone_id` - The Route53 Hosted Zone ID of the endpoint

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Overview.Replication.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Aurora.html
[3]: /docs/providers/aws/r/rds_cluster_instance.html
[4]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html
[5]: http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Limits.html#RDS_Limits.Constraints
[6]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html#UsingWithRDS.IAMDBAuth.Availability

## Timeouts

`aws_rds_cluster` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `120 minutes`) Used for Cluster creation
- `update` - (Default `120 minutes`) Used for Cluster modifications
- `delete` - (Default `120 minutes`) Used for destroying cluster. This includes
any cleanup task during the destroying process.

## Import

RDS Clusters can be imported using the `cluster_identifier`, e.g.

```
$ terraform import aws_rds_cluster.aurora_cluster aurora-prod-cluster
```
