---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster"
description: |-
    Provides details about a specific redshift cluster
---

# Data Source: aws_redshift_cluster

Provides details about a specific redshift cluster.

## Example Usage

```hcl
data "aws_redshift_cluster" "test_cluster" {
  cluster_identifier = "test-cluster"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "redshift"

  s3_configuration {
    role_arn           = aws_iam_role.firehose_role.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  redshift_configuration {
    role_arn           = aws_iam_role.firehose_role.arn
    cluster_jdbcurl    = "jdbc:redshift://${data.aws_redshift_cluster.test_cluster.endpoint}/${data.aws_redshift_cluster.test_cluster.database_name}"
    username           = "testuser"
    password           = "T3stPass"
    data_table_name    = "test-table"
    copy_options       = "delimiter '|'" # the default delimiter
    data_table_columns = "test-col"
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The cluster identifier

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `allow_version_upgrade` - Whether major version upgrades can be applied during maintenance period
* `automated_snapshot_retention_period` - The backup retention period
* `availability_zone` - The availability zone of the cluster
* `bucket_name` - The name of the S3 bucket where the log files are to be stored
* `cluster_identifier` - The cluster identifier
* `cluster_parameter_group_name` - The name of the parameter group to be associated with this cluster
* `cluster_public_key` - The public key for the cluster
* `cluster_revision_number` - The cluster revision number
* `cluster_security_groups` - The security groups associated with the cluster
* `cluster_subnet_group_name` - The name of a cluster subnet group to be associated with this cluster
* `cluster_type` - The cluster type
* `database_name` - The name of the default database in the cluster
* `elastic_ip` - The Elastic IP of the cluster
* `enable_logging` - Whether cluster logging is enabled
* `encrypted` - Whether the cluster data is encrypted
* `endpoint` - The cluster endpoint
* `enhanced_vpc_routing` - Whether enhanced VPC routing is enabled
* `iam_roles` - The IAM roles associated to the cluster
* `kms_key_id` - The KMS encryption key associated to the cluster
* `master_username` - Username for the master DB user
* `node_type` - The cluster node type
* `number_of_nodes` - The number of nodes in the cluster
* `port` - The port the cluster responds on
* `preferred_maintenance_window` - The maintenance window
* `publicly_accessible` - Whether the cluster is publicly accessible
* `s3_key_prefix` - The folder inside the S3 bucket where the log files are stored
* `tags` - The tags associated to the cluster
* `vpc_id` - The VPC Id associated with the cluster
* `vpc_security_group_ids` - The VPC security group Ids associated with the cluster
