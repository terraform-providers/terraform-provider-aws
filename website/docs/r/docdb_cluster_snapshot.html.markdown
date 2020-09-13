---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_cluster_snapshot"
description: |-
  Manages a DocDB database cluster snapshot.
---

# Resource: aws_docdb_cluster_snapshot

Manages a DocDB database cluster snapshot for DocDB clusters.

## Example Usage

```hcl
resource "aws_docdb_cluster_snapshot" "example" {
  db_cluster_identifier          = aws_docdb_cluster.example.id
  db_cluster_snapshot_identifier = "resourcetestsnapshot1234"
}
```

## Argument Reference

The following arguments are supported:

* `db_cluster_identifier` - (Required) The DocDB Cluster Identifier from which to take the snapshot.
* `db_cluster_snapshot_identifier` - (Required) The Identifier for the snapshot.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `availability_zones` - List of EC2 Availability Zones that instances in the DocDB cluster snapshot can be restored in.
* `db_cluster_snapshot_arn` - The Amazon Resource Name (ARN) for the DocDB Cluster Snapshot.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Version of the database engine for this DocDB cluster snapshot.
* `kms_key_id` - If storage_encrypted is true, the AWS KMS key identifier for the encrypted DocDB cluster snapshot.
* `port` - Port that the DocDB cluster was listening on at the time of the snapshot.
* `source_db_cluster_snapshot_identifier` - The DocDB Cluster Snapshot Arn that the DocDB Cluster Snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `storage_encrypted` - Specifies whether the DocDB cluster snapshot is encrypted.
* `status` - The status of this DocDB Cluster Snapshot.
* `vpc_id` - The VPC ID associated with the DocDB cluster snapshot.

## Timeouts

`aws_docdb_cluster_snapshot` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `20m`) How long to wait for the snapshot to be available.

## Import

`aws_docdb_cluster_snapshot` can be imported by using the cluster snapshot identifier, e.g.

```
$ terraform import aws_docdb_cluster_snapshot.example my-cluster-snapshot
```
