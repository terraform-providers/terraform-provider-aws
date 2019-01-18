---
layout: "aws"
page_title: "AWS: aws_db_cluster_snapshot"
sidebar_current: "docs-aws-datasource-db-cluster-snapshot"
description: |-
  Get information on a DB Cluster Snapshot.
---

# Data Source: aws_db_cluster_snapshot

Use this data source to get information about a DB Cluster Snapshot for use when provisioning DB clusters.

~> **NOTE:** This data source does not apply to snapshots created on DB Instances. 
See the [`aws_db_snapshot` data source](/docs/providers/aws/d/db_snapshot.html) for DB Instance snapshots.

## Example Usage

```hcl
data "aws_db_cluster_snapshot" "development_final_snapshot" {
  db_cluster_identifier = "development_cluster"
  most_recent           = true
}

# Use the last snapshot of the dev database before it was destroyed to create
# a new dev database.
resource "aws_rds_cluster" "aurora" {
  cluster_identifier   = "development_cluster"
  snapshot_identifier  = "${data.aws_db_cluster_snapshot.development_final_snapshot.id}"
  db_subnet_group_name = "my_db_subnet_group"

  lifecycle {
    ignore_changes = ["snapshot_identifier"]
  }
}

resource "aws_rds_cluster_instance" "aurora" {
  cluster_identifier   = "${aws_rds_cluster.aurora.id}"
  instance_class       = "db.t2.small"
  db_subnet_group_name = "my_db_subnet_group"
}
```

## Argument Reference

The following arguments are supported:

* `most_recent` - (Optional) If more than one result is returned, use the most recent Snapshot.

* `db_cluster_identifier` - (Optional) Returns the list of snapshots created by the specific db_cluster

* `db_cluster_snapshot_identifier` - (Optional) Returns information on a specific snapshot_id.

* `snapshot_type` - (Optional) The type of snapshots to be returned. If you don't specify a SnapshotType
value, then both automated and manual DB cluster snapshots are returned. Shared and public DB Cluster Snapshots are not
included in the returned results by default. Possible values are, `automated`, `manual`, `shared` and `public`.

* `include_shared` - (Optional) Set this value to true to include shared manual DB Cluster Snapshots from other
AWS accounts that this AWS account has been given permission to copy or restore, otherwise set this value to false.
The default is `false`.

* `include_public` - (Optional) Set this value to true to include manual DB Cluster Snapshots that are public and can be
copied or restored by any AWS account, otherwise set this value to false. The default is `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `allocated_storage` - Specifies the allocated storage size in gigabytes (GB).
* `availability_zones` - List of EC2 Availability Zones that instances in the DB cluster snapshot can be restored in.
* `db_cluster_identifier` - Specifies the DB cluster identifier of the DB cluster that this DB cluster snapshot was created from.
* `db_cluster_snapshot_arn` - The Amazon Resource Name (ARN) for the DB Cluster Snapshot.
* `engine_version` - Version of the database engine for this DB cluster snapshot.
* `engine` - Specifies the name of the database engine.
* `id` - The snapshot ID.
* `kms_key_id` - If storage_encrypted is true, the AWS KMS key identifier for the encrypted DB cluster snapshot.
* `license_model` - License model information for the restored DB cluster.
* `port` - Port that the DB cluster was listening on at the time of the snapshot.
* `snapshot_create_time` - Time when the snapshot was taken, in Universal Coordinated Time (UTC).
* `source_db_cluster_snapshot_identifier` - The DB Cluster Snapshot Arn that the DB Cluster Snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `status` - The status of this DB Cluster Snapshot.
* `storage_encrypted` - Specifies whether the DB cluster snapshot is encrypted.
* `vpc_id` - The VPC ID associated with the DB cluster snapshot.
