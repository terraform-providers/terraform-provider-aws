---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_global_cluster"
description: |-
  Manages an RDS Global Cluster
---

# Resource: aws_rds_global_cluster

Manages an RDS Global Cluster, which is an Aurora global database spread across multiple regions. The global database contains a single primary cluster with read-write capability, and a read-only secondary cluster that receives data from the primary cluster through high-speed replication performed by the Aurora storage subsystem.

More information about Aurora global databases can be found in the [Aurora User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html#aurora-global-database-creating).

## Example Usage

### New Global Cluster

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
  provider = aws.primary

  global_cluster_identifier = "example"
}

resource "aws_rds_cluster" "primary" {
  provider = aws.primary

  # ... other configuration ...
  global_cluster_identifier = aws_rds_global_cluster.example.id
}

resource "aws_rds_cluster_instance" "primary" {
  provider = aws.primary

  # ... other configuration ...
  cluster_identifier = aws_rds_cluster.primary.id
}

resource "aws_rds_cluster" "secondary" {
  depends_on = [aws_rds_cluster_instance.primary]
  provider   = aws.secondary

  # ... other configuration ...
  global_cluster_identifier = aws_rds_global_cluster.example.id
}

resource "aws_rds_cluster_instance" "secondary" {
  provider = aws.secondary

  # ... other configuration ...
  cluster_identifier = aws_rds_cluster.secondary.id
}
```

### New Global Cluster From Existing DB Cluster

```hcl
resource "aws_rds_cluster" "example" {
  # ... other configuration ...

  # NOTE: Using this DB Cluster to create a Global Cluster, the
  # global_cluster_identifier attribute will become populated and
  # Terraform will begin showing it as a difference. Do not configure:
  # global_cluster_identifier = aws_rds_global_cluster.example.id
  # as it creates a circular reference. Use ignore_changes instead.
  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_rds_global_cluster" "example" {
  force_destroy                = true
  global_cluster_identifier    = "example"
  source_db_cluster_identifier = aws_rds_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `global_cluster_identifier` - (Required, Forces new resources) The global cluster identifier.
* `database_name` - (Optional, Forces new resources) Name for an automatically created database on cluster creation.
* `deletion_protection` - (Optional) If the Global Cluster should have deletion protection enabled. The database can't be deleted when this value is set to `true`. The default is `false`.
* `engine` - (Optional, Forces new resources) Name of the database engine to be used for this DB cluster. Terraform will only perform drift detection if a configuration value is provided. Valid values: `aurora`, `aurora-mysql`, `aurora-postgresql`. Defaults to `aurora`. Conflicts with `source_db_cluster_identifier`.
* `engine_version` - (Optional, Forces new resources) Engine version of the Aurora global database.
    * **NOTE:** When the engine is set to `aurora-mysql`, an engine version compatible with global database is required. The earliest available version is `5.7.mysql_aurora.2.06.0`.
* `force_destroy` - (Optional) Enable to remove DB Cluster members from Global Cluster on destroy. Required with `source_db_cluster_identifier`.
* `source_db_cluster_identifier` - (Optional) Amazon Resource Name (ARN) to use as the primary DB Cluster of the Global Cluster on creation. Terraform cannot perform drift detection of this value.
* `storage_encrypted` - (Optional, Forces new resources) Specifies whether the DB cluster is encrypted. The default is `false` unless `source_db_cluster_identifier` is specified and encrypted. Terraform will only perform drift detection if a configuration value is provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - RDS Global Cluster Amazon Resource Name (ARN)
* `global_cluster_members` - Set of objects containing Global Cluster members.
    * `db_cluster_arn` - Amazon Resource Name (ARN) of member DB Cluster
    * `is_writer` - Whether the member is the primary DB Cluster
* `global_cluster_resource_id` - AWS Region-unique, immutable identifier for the global database cluster. This identifier is found in AWS CloudTrail log entries whenever the AWS KMS key for the DB cluster is accessed
* `id` - RDS Global Cluster identifier

## Import

`aws_rds_global_cluster` can be imported by using the RDS Global Cluster identifier, e.g.

```
$ terraform import aws_rds_global_cluster.example example
```

Certain resource arguments, like `force_destroy`, only exist within Terraform. If the argument is set in the Terraform configuration on an imported resource, Terraform will show a difference on the first plan after import to update the state value. This change is safe to apply immediately so the state matches the desired configuration.

Certain resource arguments, like `source_db_cluster_identifier`, do not have an API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.

```hcl
resource "aws_rds_global_cluster" "example" {
  # ... other configuration ...

  # There is no API for reading source_db_cluster_identifier
  lifecycle {
    ignore_changes = [source_db_cluster_identifier]
  }
}
```
