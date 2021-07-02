---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_cluster_activity_stream"
description: |-
  Manages RDS Aurora Cluster Database Activity Streams
---

# Resource: aws_rds_cluster_activity_stream

Manages RDS Aurora Cluster Database Activity Streams.

Database Activity Streams have some limits and requirements, You can refer to the [User Guide][1].

~> **Note:** `apply_immediately` always is true, cannot be modified.
Because when apply_immediately=false, terraform cannot get activity stream associated attributes.

~> **Note:** This resource depends on having one `aws_rds_cluster_instance` created.
To avoid race conditions when all resources are being created together, you need to add explicit resource
references using the [resource `depends_on` meta-argument](/docs/configuration/resources.html#depends_on-explicit-resource-dependencies).


## Example Usage

```terraform
resource "aws_rds_cluster" "default" {
  cluster_identifier = "aurora-cluster-demo"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "mustbeeightcharaters"
  engine             = "aurora-postgresql"
  engine_version     = "10.11"
}

resource "aws_rds_cluster_instance" "default" {
  identifier         = "aurora-instance-demo"
  cluster_identifier = aws_rds_cluster.default.cluster_identifier
  engine             = aws_rds_cluster.default.engine
  instance_class     = "db.r5.large"
}

resource "aws_kms_key" "default" {
  description = "aws kms key"
}

resource "aws_rds_cluster_activity_stream" "default" {
  resource_arn      = aws_rds_cluster.default.arn
  mode              = "async"
  kms_key_id        = aws_kms_key.default.key_id
  apply_immediately = true

  depends_on = [aws_rds_cluster_instance.default]
}
```


## Argument Reference

For more detailed documentation about each argument, refer to
the [AWS official documentation](https://docs.aws.amazon.com/cli/latest/reference/rds/start-activity-stream.html).

The following arguments are supported:

* `resource_arn` - (Required, Forces new resources) The Amazon Resource Name (ARN) of the DB cluster.
* `mode` - (Required, Forces new resources) Specifies the mode of the database activity stream. One of: `sync` , `async` .
* `kms_key_id` - (Required, Forces new resources) The AWS KMS key identifier used for encrypting messages in the database activity stream.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the DB cluster.
* `kinesis_stream_name` - The name of the Amazon Kinesis data stream to be used for the database activity stream.


## Import

RDS Aurora Cluster Database Activity Streams can be imported using the `resource_arn`, e.g.

```
$ terraform import aws_rds_cluster_activity_stream.default arn:aws:rds:us-west-2:123456789012:cluster:aurora-cluster-demo
```

[1]: https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/DBActivityStreams.html
[2]: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_UpgradeDBInstance.Maintenance.html
