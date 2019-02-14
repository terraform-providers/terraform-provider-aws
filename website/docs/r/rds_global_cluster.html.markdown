---
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
sidebar_current: "docs-aws-resource-ec2-transit-gateway-x"
description: |-
  Manages a RDS Global Cluster
---

# aws_rds_global_cluster

Manages a RDS Global Cluster, which is an Aurora global database spread across multiple regions. The global database contains a single primary cluster with read-write capability, and a read-only secondary cluster that receives data from the primary cluster through high-speed replication performed by the Aurora storage subsystem.

More information about Aurora global databases can be found in the [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html#aurora-global-database-creating).

~> **NOTE:** RDS only supports the `aurora` engine (MySQL 5.6 compatible) for Global Clusters at this time.

## Example Usage

```hcl
provider "aws" {
  alias  = "primary"
  region = "us-east-2"
}

provider "aws" {
  alias  = "secondary"
  region = "us-west-2"
}

resource "aws_rds_global_cluster" "example" {
  provider = "aws.primary"

  global_cluster_identifier = "example"
}

resource "aws_rds_cluster" "primary" {
  provider = "aws.primary"

  # ... other configuration ...
  engine_mode               = "global"
  global_cluster_identifier = "${aws_rds_global_cluster.example.id}"
}

resource "aws_rds_cluster_instance" "primary" {
  provider = "aws.primary"

  # ... other configuration ...
  cluster_identifier = "${aws_rds_cluster.primary.id}"
}

resource "aws_rds_cluster" "secondary" {
  depends_on = ["aws_rds_cluster_instance.primary"]
  provider   = "aws.secondary"

  # ... other configuration ...
  engine_mode               = "global"
  global_cluster_identifier = "${aws_rds_global_cluster.example.id}"
}

resource "aws_rds_cluster_instance" "secondary" {
  provider = "aws.secondary"

  # ... other configuration ...
  cluster_identifier = "${aws_rds_cluster.secondary.id}"
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Optional) Name for an automatically created database on cluster creation.
* `deletion_protection` - (Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `engine` - (Optional) Name of the database engine to be used for this DB cluster. Valid values: `aurora`. Defaults to `aurora`.
* `engine_version` - (Optional) Engine version of the Aurora global database.
* `storage_encrypted` - (Optional) Specifies whether the DB cluster is encrypted. The default is `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - RDS Global Cluster Amazon Resource Name (ARN)
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed
* `id` - RDS Global Cluster identifier

## Import

`aws_rds_global_cluster` can be imported by using the RDS Global Cluster identifier, e.g.

```
$ terraform import aws_rds_global_cluster.example example
```
