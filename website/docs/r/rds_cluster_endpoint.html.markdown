---
layout: "aws"
page_title: "AWS: aws_rds_cluster_endpoint"
sidebar_current: "docs-aws-resource-rds-cluster-endpoint"
description: |-
  Manages a RDS Aurora Cluster Endpoint
---

# aws_rds_cluster_endpoint

Manages a RDS Aurora Cluster Endpoint.
You can refer to the [User Guide][1].


## Example Usage

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

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately       = true
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  identifier              = "test1"
  instance_class          = "db.t2.small"
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately       = true
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  identifier              = "test2"
  instance_class          = "db.t2.small"
}

resource "aws_rds_cluster_instance" "test3" {
  apply_immediately       = true
  cluster_identifier      = "${aws_rds_cluster.default.id}"
  identifier              = "test3"
  instance_class          = "db.t2.small"
}

resource "aws_rds_cluster_endpoint" "eligible" {
  cluster_identifier          = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "reader"
  custom_endpoint_type        = "READER"
  
  excluded_members = [
    "${aws_rds_cluster_instance.test1.id}",
    "${aws_rds_cluster_instance.test2.id}",
  ]
}

resource "aws_rds_cluster_endpoint" "static" {
  cluster_identifier          = "${aws_rds_cluster.default.id}"
  cluster_endpoint_identifier = "static"
  custom_endpoint_type        = "READER"
  
  static_members = [
    "${aws_rds_cluster_instance.test1.id}",
    "${aws_rds_cluster_instance.test3.id}",
  ]
}
```

## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/create-db-cluster-endpoint.html).

The following arguments are supported:

* `cluster_identifier` - (Required, Forces new resources) The cluster identifier.
* `cluster_endpoint_identifier` - (Required, Forces new resources) The identifier to use for the new endpoint. This parameter is stored as a lowercase string.
* `custom_endpoint_type` - (Required) The type of the endpoint. One of: READER , ANY .
* `static_members` - (Optional) List of DB instance identifiers that are part of the custom endpoint group. Conflicts with `excluded_members`.
* `excluded_members` - (Optional) List of DB instance identifiers that aren't part of the custom endpoint group. All other eligible instances are reachable through the custom endpoint. Only relevant if the list of static members is empty. Conflicts with `static_members`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of cluster
* `id` - The RDS Cluster Endpoint Identifier
* `endpoint` - A custom endpoint for the Aurora cluster


## Import

RDS Clusters Endpoint can be imported using the `cluster_endpoint_identifier`, e.g.

```
$ terraform import aws_rds_cluster_endpoint.custom_reader aurora-prod-cluster-custom-reader
```

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Overview.Endpoints.html#Aurora.Endpoints.Cluster
