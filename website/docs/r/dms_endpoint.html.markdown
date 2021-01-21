---
subcategory: "Database Migration Service (DMS)"
layout: "aws"
page_title: "AWS: aws_dms_endpoint"
description: |-
  Provides a DMS (Data Migration Service) endpoint resource.
---

# Resource: aws_dms_endpoint

Provides a DMS (Data Migration Service) endpoint resource. DMS endpoints can be created, updated, deleted, and imported.

~> **Note:** All arguments including the password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```hcl
# Create a new endpoint
resource "aws_dms_endpoint" "test" {
  certificate_arn             = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
  database_name               = "test"
  endpoint_id                 = "test-dms-endpoint-tf"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = ""
  kms_key_arn                 = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  password                    = "test"
  port                        = 3306
  server_name                 = "test"
  ssl_mode                    = "none"

  tags = {
    Name = "test"
  }

  username = "test"
}
```

## Argument Reference

The following arguments are supported:

* `certificate_arn` - (Optional, Default: empty string) The Amazon Resource Name (ARN) for the certificate.
* `database_name` - (Optional) The name of the endpoint database.
* `elasticsearch_settings` - (Optional) Configuration block with Elasticsearch settings. Detailed below.
* `endpoint_id` - (Required) The database endpoint identifier.

    - Must contain from 1 to 255 alphanumeric characters or hyphens.
    - Must begin with a letter
    - Must contain only ASCII letters, digits, and hyphens
    - Must not end with a hyphen
    - Must not contain two consecutive hyphens

* `endpoint_type` - (Required) The type of endpoint. Can be one of `source | target`.
* `engine_name` - (Required) The type of engine for the endpoint. Can be one of `aurora | aurora-postgresql| azuredb | db2 | docdb | dynamodb | elasticsearch | kafka | kinesis | mariadb | mongodb | mysql | oracle | postgres | redshift | s3 | sqlserver | sybase`.
* `extra_connection_attributes` - (Optional) Additional attributes associated with the connection. For available attributes see [Using Extra Connection Attributes with AWS Database Migration Service](http://docs.aws.amazon.com/dms/latest/userguide/CHAP_Introduction.ConnectionAttributes.html).
* `kafka_settings` - (Optional) Configuration block with Kafka settings. Detailed below.
* `kinesis_settings` - (Optional) Configuration block with Kinesis settings. Detailed below.
* `kms_key_arn` - (Required when `engine_name` is `mongodb`, optional otherwise) The Amazon Resource Name (ARN) for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region.
* `mongodb_settings` - (Optional) Configuration block with MongoDB settings. Detailed below.
* `password` - (Optional) The password to be used to login to the endpoint database.
* `port` - (Optional) The port used by the endpoint database.
* `s3_settings` - (Optional) Configuration block with S3 settings. Detailed below.
* `server_name` - (Optional) The host name of the server.
* `service_access_role` - (Optional) The Amazon Resource Name (ARN) used by the service access IAM role for dynamodb endpoints.
* `ssl_mode` - (Optional, Default: none) The SSL mode to use for the connection. Can be one of `none | require | verify-ca | verify-full`
* `tags` - (Optional) A map of tags to assign to the resource.
* `username` - (Optional) The user name to be used to login to the endpoint database.

### elasticsearch_settings Arguments

-> Additional information can be found in the [Using Amazon Elasticsearch Service as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Elasticsearch.html).

The `elasticsearch_settings` configuration block supports the following arguments:

* `endpoint_uri` - (Required) Endpoint for the Elasticsearch cluster.
* `error_retry_duration` - (Optional) Maximum number of seconds for which DMS retries failed API requests to the Elasticsearch cluster. Defaults to `300`.
* `full_load_error_percentage` - (Optional) Maximum percentage of records that can fail to be written before a full load operation stops. Defaults to `10`.
* `service_access_role_arn` - (Required) Amazon Resource Name (ARN) of the IAM Role with permissions to write to the Elasticsearch cluster.

### kafka_settings Arguments

-> Additional information can be found in the [Using Apache Kafka as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kafka.html).

The `kafka_settings` configuration block supports the following arguments:

* `broker` - (Required) Kafka broker location. Specify in the form broker-hostname-or-ip:port.
* `topic` - (Optional) Kafka topic for migration. Defaults to `kafka-default-topic`.

### kinesis_settings Arguments

-> Additional information can be found in the [Using Amazon Kinesis Data Streams as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.Kinesis.html).

The `kinesis_settings` configuration block supports the following arguments:

* `message_format` - (Optional) Output format for the records created. Defaults to `json`. Valid values are `json` and `json_unformatted` (a single line with no tab).
* `service_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role with permissions to write to the Kinesis data stream.
* `stream_arn` - (Optional) Amazon Resource Name (ARN) of the Kinesis data stream.

### mongodb_settings Arguments

-> Additional information can be found in the [Using MongoDB as a Source for AWS DMS documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.MongoDB.html).

The `mongodb_settings` configuration block supports the following arguments:

* `auth_mechanism` - (Optional) Authentication mechanism to access the MongoDB source endpoint. Defaults to `default`.
* `auth_source` - (Optional) Authentication database name. Not used when `auth_type` is `no`. Defaults to `admin`.
* `auth_type` - (Optional) Authentication type to access the MongoDB source endpoint. Defaults to `password`.
* `docs_to_investigate` - (Optional) Number of documents to preview to determine the document organization. Use this setting when `nesting_level` is set to `one`. Defaults to `1000`.
* `extract_doc_id` - (Optional) Document ID. Use this setting when `nesting_level` is set to `none`. Defaults to `false`.
* `nesting_level` - (Optional) Specifies either document or table mode. Defaults to `none`. Valid values are `one` (table mode) and `none` (document mode).

### s3_settings Arguments

-> Additional information can be found in the [Using Amazon S3 as a Source for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.S3.html) and [Using Amazon S3 as a Target for AWS Database Migration Service documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Target.S3.html).

The `s3_settings` configuration block supports the following arguments:

* `bucket_folder` - (Optional) S3 Bucket Object prefix.
* `bucket_name` - (Optional) S3 Bucket name.
* `compression_type` - (Optional) Set to compress target files. Defaults to `NONE`. Valid values are `GZIP` and `NONE`.
* `csv_delimiter` - (Optional) Delimiter used to separate columns in the source files. Defaults to `,`.
* `csv_row_delimiter` - (Optional) Delimiter used to separate rows in the source files. Defaults to `\n`.
* `date_partition_enabled` - (Optional) Partition S3 bucket folders based on transaction commit dates. Defaults to `false`.
* `external_table_definition` - (Optional) JSON document that describes how AWS DMS should interpret the data.
* `service_access_role_arn` - (Optional) Amazon Resource Name (ARN) of the IAM Role with permissions to read from or write to the S3 Bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint_arn` - The Amazon Resource Name (ARN) for the endpoint.

## Import

Endpoints can be imported using the `endpoint_id`, e.g.

```
$ terraform import aws_dms_endpoint.test test-dms-endpoint-tf
```
