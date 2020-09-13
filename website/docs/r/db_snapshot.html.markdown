---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_snapshot"
description: |-
  Manages an RDS database instance snapshot.
---

# Resource: aws_db_snapshot

Manages an RDS database instance snapshot. For managing RDS database cluster snapshots, see the [`aws_db_cluster_snapshot` resource](/docs/providers/aws/r/db_cluster_snapshot.html).

## Example Usage

```hcl
resource "aws_db_instance" "bar" {
  allocated_storage = 10
  engine            = "MySQL"
  engine_version    = "5.6.21"
  instance_class    = "db.t2.micro"
  name              = "baz"
  password          = "barbarbarbar"
  username          = "foo"

  maintenance_window      = "Fri:09:00-Fri:09:30"
  backup_retention_period = 0
  parameter_group_name    = "default.mysql5.6"
}

resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.bar.id
  db_snapshot_identifier = "testsnapshot1234"
}
```

## Argument Reference

The following arguments are supported:

* `db_instance_identifier` - (Required) The DB Instance Identifier from which to take the snapshot.
* `db_snapshot_identifier` - (Required) The Identifier for the snapshot.
* `tags` - (Optional) Key-value map of resource tags


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `allocated_storage` - Specifies the allocated storage size in gigabytes (GB).
* `availability_zone` - Specifies the name of the Availability Zone the DB instance was located in at the time of the DB snapshot.
* `db_snapshot_arn` - The Amazon Resource Name (ARN) for the DB snapshot.
* `encrypted` - Specifies whether the DB snapshot is encrypted.
* `engine` - Specifies the name of the database engine.
* `engine_version` - Specifies the version of the database engine.
* `iops` - Specifies the Provisioned IOPS (I/O operations per second) value of the DB instance at the time of the snapshot.
* `kms_key_id` - The ARN for the KMS encryption key.
* `license_model` - License model information for the restored DB instance.
* `option_group_name` - Provides the option group name for the DB snapshot.
* `source_db_snapshot_identifier` - The DB snapshot Arn that the DB snapshot was copied from. It only has value in case of cross customer or cross region copy.
* `source_region` - The region that the DB snapshot was created in or copied from.
* `status` - Specifies the status of this DB snapshot.
* `storage_type` - Specifies the storage type associated with DB snapshot.
* `vpc_id` - Specifies the storage type associated with DB snapshot.

## Timeouts

`aws_db_snapshot` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `read` - (Default `20 minutes`)  Length of time to wait for the snapshot to become available

## Import

`aws_db_snapshot` can be imported by using the snapshot identifier, e.g.

```
$ terraform import aws_db_snapshot.example my-snapshot
```
