---
layout: "aws"
page_title: "AWS: aws_kinesis_firehose_delivery_stream"
sidebar_current: "docs-aws-resource-kinesis-firehose-delivery-stream"
description: |-
  Provides a AWS Kinesis Firehose Delivery Stream
---

# aws_kinesis_firehose_delivery_stream

Provides a Kinesis Firehose Delivery Stream resource. Amazon Kinesis Firehose is a fully managed, elastic service to easily deliver real-time data streams to destinations such as Amazon S3 and Amazon Redshift.

For more details, see the [Amazon Kinesis Firehose Documentation][1].

## Example Usage

### Extended S3 Destination

```hcl
resource "aws_kinesis_firehose_delivery_stream" "extended_s3_stream" {
  name        = "terraform-kinesis-firehose-extended-s3-test-stream"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = "${aws_iam_role.firehose_role.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket"
  acl    = "private"
}

resource "aws_iam_role" "firehose_role" {
  name = "firehose_test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda_iam" {
  name = "lambda_iam"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_processor" {
  filename      = "lambda.zip"
  function_name = "firehose_lambda_processor"
  role          = "${aws_iam_role.lambda_iam.arn}"
  handler       = "exports.handler"
  runtime       = "nodejs8.10"
}
```

### S3 Destination

```hcl
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket"
  acl    = "private"
}

resource "aws_iam_role" "firehose_role" {
  name = "firehose_test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "s3"

  s3_configuration {
    role_arn   = "${aws_iam_role.firehose_role.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}
```

### Redshift Destination

```hcl
resource "aws_redshift_cluster" "test_cluster" {
  cluster_identifier = "tf-redshift-cluster-%d"
  database_name      = "test"
  master_username    = "testuser"
  master_password    = "T3stPass"
  node_type          = "dc1.large"
  cluster_type       = "single-node"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "redshift"

  s3_configuration {
    role_arn           = "${aws_iam_role.firehose_role.arn}"
    bucket_arn         = "${aws_s3_bucket.bucket.arn}"
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  redshift_configuration {
    role_arn           = "${aws_iam_role.firehose_role.arn}"
    cluster_jdbcurl    = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username           = "testuser"
    password           = "T3stPass"
    data_table_name    = "test-table"
    copy_options       = "delimiter '|'" # the default delimiter
    data_table_columns = "test-col"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn           = "${aws_iam_role.firehose_role.arn}"
      bucket_arn         = "${aws_s3_bucket.bucket.arn}"
      buffer_size        = 15
      buffer_interval    = 300
      compression_format = "GZIP"
    }
  }
}
```

### Elasticsearch Destination

```hcl
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "firehose-es-test"
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "elasticsearch"

  s3_configuration {
    role_arn           = "${aws_iam_role.firehose_role.arn}"
    bucket_arn         = "${aws_s3_bucket.bucket.arn}"
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  elasticsearch_configuration {
    domain_arn = "${aws_elasticsearch_domain.test_cluster.arn}"
    role_arn   = "${aws_iam_role.firehose_role.arn}"
    index_name = "test"
    type_name  = "test"

    processing_configuration {
      enabled = "true"

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_processor.arn}:$LATEST"
        }
      }
    }
  }
}
```


### Splunk Destination

```hcl
resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = "terraform-kinesis-firehose-test-stream"
  destination = "splunk"

  s3_configuration {
    role_arn           = "${aws_iam_role.firehose.arn}"
    bucket_arn         = "${aws_s3_bucket.bucket.arn}"
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  splunk_configuration {
    hec_endpoint               = "https://http-inputs-mydomain.splunkcloud.com:443"
    hec_token                  = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
    hec_acknowledgment_timeout = 600
    hec_endpoint_type          = "Event"
    s3_backup_mode             = "FailedEventsOnly"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name to identify the stream. This is unique to the
AWS account and region the Stream is created in.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `kinesis_source_configuration` - (Optional) Allows the ability to specify the kinesis stream that is used as the source of the firehose delivery stream.
* `destination` – (Required) This is the destination to where the data is delivered. The only options are `s3` (Deprecated, use `extended_s3` instead), `extended_s3`, `redshift`, `elasticsearch`, and `splunk`.
* `s3_configuration` - (Optional, Deprecated, see/use `extended_s3_configuration` unless `destination` is `redshift`) Configuration options for the s3 destination (or the intermediate bucket if the destination
is redshift). More details are given below.
* `extended_s3_configuration` - (Optional, only Required when `destination` is `extended_s3`) Enhanced configuration options for the s3 destination. More details are given below.
* `redshift_configuration` - (Optional) Configuration options if redshift is the destination.
Using `redshift_configuration` requires the user to also specify a
`s3_configuration` block. More details are given below.

The `kinesis_source_configuration` object supports the following:

* `kinesis_stream_arn` (Required) The kinesis stream used as the source of the firehose delivery stream.
* `role_arn` (Required) The ARN of the role that provides access to the source Kinesis stream.

The `s3_configuration` object supports the following:

* `role_arn` - (Required) The ARN of the AWS credentials.
* `bucket_arn` - (Required) The ARN of the S3 bucket
* `prefix` - (Optional) The "YYYY/MM/DD/HH" time format prefix is automatically used for delivered S3 files. You can specify an extra prefix to be added in front of the time format prefix. Note that if the prefix ends with a slash, it appears as a folder in the S3 bucket
* `buffer_size` - (Optional) Buffer incoming data to the specified size, in MBs, before delivering it to the destination. The default value is 5.
                                We recommend setting SizeInMBs to a value greater than the amount of data you typically ingest into the delivery stream in 10 seconds. For example, if you typically ingest data at 1 MB/sec set SizeInMBs to be 10 MB or higher.
* `buffer_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds, before delivering it to the destination. The default value is 300.
* `compression_format` - (Optional) The compression format. If no value is specified, the default is UNCOMPRESSED. Other supported values are GZIP, ZIP & Snappy. If the destination is redshift you cannot use ZIP or Snappy.
* `kms_key_arn` - (Optional) Specifies the KMS key ARN the stream will use to encrypt data. If not set, no encryption will
be used.
* `cloudwatch_logging_options` - (Optional) The CloudWatch Logging Options for the delivery stream. More details are given below

The `extended_s3_configuration` object supports the same fields from `s3_configuration` as well as the following:

* `data_format_conversion_configuration` - (Optional) Nested argument for the serializer, deserializer, and schema for converting data from the JSON format to the Parquet or ORC format before writing it to Amazon S3. More details given below.
* `error_output_prefix` - (Optional) Prefix added to failed records before writing them to S3. This prefix appears immediately following the bucket name.
* `processing_configuration` - (Optional) The data processing configuration.  More details are given below.
* `s3_backup_mode` - (Optional) The Amazon S3 backup mode.  Valid values are `Disabled` and `Enabled`.  Default value is `Disabled`.
* `s3_backup_configuration` - (Optional) The configuration for backup in Amazon S3. Required if `s3_backup_mode` is `Enabled`. Supports the same fields as `s3_configuration` object.

The `redshift_configuration` object supports the following:

* `cluster_jdbcurl` - (Required) The jdbcurl of the redshift cluster.
* `username` - (Required) The username that the firehose delivery stream will assume. It is strongly recommended that the username and password provided is used exclusively for Amazon Kinesis Firehose purposes, and that the permissions for the account are restricted for Amazon Redshift INSERT permissions.
* `password` - (Required) The password for the username above.
* `retry_duration` - (Optional) The length of time during which Firehose retries delivery after a failure, starting from the initial request and including the first attempt. The default value is 3600 seconds (60 minutes). Firehose does not retry if the value of DurationInSeconds is 0 (zero) or if the first delivery attempt takes longer than the current value.
* `role_arn` - (Required) The arn of the role the stream assumes.
* `s3_backup_mode` - (Optional) The Amazon S3 backup mode.  Valid values are `Disabled` and `Enabled`.  Default value is `Disabled`.
* `s3_backup_configuration` - (Optional) The configuration for backup in Amazon S3. Required if `s3_backup_mode` is `Enabled`. Supports the same fields as `s3_configuration` object.
* `data_table_name` - (Required) The name of the table in the redshift cluster that the s3 bucket will copy to.
* `copy_options` - (Optional) Copy options for copying the data from the s3 intermediate bucket into redshift, for example to change the default delimiter. For valid values, see the [AWS documentation](http://docs.aws.amazon.com/firehose/latest/APIReference/API_CopyCommand.html)
* `data_table_columns` - (Optional) The data table columns that will be targeted by the copy command.
* `cloudwatch_logging_options` - (Optional) The CloudWatch Logging Options for the delivery stream. More details are given below
* `processing_configuration` - (Optional) The data processing configuration.  More details are given below.

The `elasticsearch_configuration` object supports the following:

* `buffering_interval` - (Optional) Buffer incoming data for the specified period of time, in seconds between 60 to 900, before delivering it to the destination.  The default value is 300s.
* `buffering_size` - (Optional) Buffer incoming data to the specified size, in MBs between 1 to 100, before delivering it to the destination.  The default value is 5MB.
* `domain_arn` - (Required) The ARN of the Amazon ES domain.  The IAM role must have permission for `DescribeElasticsearchDomain`, `DescribeElasticsearchDomains`, and `DescribeElasticsearchDomainConfig` after assuming `RoleARN`.  The pattern needs to be `arn:.*`.
* `index_name` - (Required) The Elasticsearch index name.
* `index_rotation_period` - (Optional) The Elasticsearch index rotation period.  Index rotation appends a timestamp to the IndexName to facilitate expiration of old data.  Valid values are `NoRotation`, `OneHour`, `OneDay`, `OneWeek`, and `OneMonth`.  The default value is `OneDay`.
* `retry_duration` - (Optional) After an initial failure to deliver to Amazon Elasticsearch, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `role_arn` - (Required) The ARN of the IAM role to be assumed by Firehose for calling the Amazon ES Configuration API and for indexing documents.  The pattern needs to be `arn:.*`.
* `s3_backup_mode` - (Optional) Defines how documents should be delivered to Amazon S3.  Valid values are `FailedDocumentsOnly` and `AllDocuments`.  Default value is `FailedDocumentsOnly`.
* `type_name` - (Required) The Elasticsearch type name with maximum length of 100 characters.
* `cloudwatch_logging_options` - (Optional) The CloudWatch Logging Options for the delivery stream. More details are given below
* `processing_configuration` - (Optional) The data processing configuration.  More details are given below.

The `splunk_configuration` objects supports the following:

* `hec_acknowledgment_timeout` - (Optional) The amount of time, in seconds between 180 and 600, that Kinesis Firehose waits to receive an acknowledgment from Splunk after it sends it data.
* `hec_endpoint` - (Required) The HTTP Event Collector (HEC) endpoint to which Kinesis Firehose sends your data.
* `hec_endpoint_type` - (Optional) The HEC endpoint type. Valid values are `Raw` or `Event`. The default value is `Raw`.
* `hec_token` - The GUID that you obtain from your Splunk cluster when you create a new HEC endpoint.
* `s3_backup_mode` - (Optional) Defines how documents should be delivered to Amazon S3.  Valid values are `FailedEventsOnly` and `AllEvents`.  Default value is `FailedEventsOnly`.
* `retry_duration` - (Optional) After an initial failure to deliver to Amazon Elasticsearch, the total amount of time, in seconds between 0 to 7200, during which Firehose re-attempts delivery (including the first attempt).  After this time has elapsed, the failed documents are written to Amazon S3.  The default value is 300s.  There will be no retry if the value is 0.
* `cloudwatch_logging_options` - (Optional) The CloudWatch Logging Options for the delivery stream. More details are given below.
* `processing_configuration` - (Optional) The data processing configuration.  More details are given below.

The `cloudwatch_logging_options` object supports the following:

* `enabled` - (Optional) Enables or disables the logging. Defaults to `false`.
* `log_group_name` - (Optional) The CloudWatch group name for logging. This value is required if `enabled` is true.
* `log_stream_name` - (Optional) The CloudWatch log stream name for logging. This value is required if `enabled` is true.

The `processing_configuration` object supports the following:

* `enabled` - (Optional) Enables or disables data processing.
* `processors` - (Optional) Array of data processors. More details are given below

The `processors` array objects support the following:

* `type` - (Required) The type of processor. Valid Values: `Lambda`
* `parameters` - (Optional) Array of processor parameters. More details are given below

The `parameters` array objects support the following:

* `parameter_name` - (Required) Parameter name. Valid Values: `LambdaArn`, `NumberOfRetries`, `RoleArn`, `BufferSizeInMBs`, `BufferIntervalInSeconds`
* `parameter_value` - (Required) Parameter value. Must be between 1 and 512 length (inclusive). When providing a Lambda ARN, you should specify the resource version as well.

### data_format_conversion_configuration

Example:

```hcl
resource "aws_kinesis_firehose_delivery_stream" "example" {
  # ... other configuration ...
  extended_s3_configuration {
    # Must be at least 64
    buffer_size = 128

    # ... other configuration ...
    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = "${aws_glue_catalog_table.example.database_name}"
        role_arn      = "${aws_iam_role.example.arn}"
        table_name    = "${aws_glue_catalog_table.example.name}"
      }
    }
  }
}
```

* `input_format_configuration` - (Required) Nested argument that specifies the deserializer that you want Kinesis Data Firehose to use to convert the format of your data from JSON. More details below.
* `output_format_configuration` - (Required) Nested argument that specifies the serializer that you want Kinesis Data Firehose to use to convert the format of your data to the Parquet or ORC format. More details below.
* `schema_configuration` - (Required) Nested argument that specifies the AWS Glue Data Catalog table that contains the column information. More details below.
* `enabled` - (Optional) Defaults to `true`. Set it to `false` if you want to disable format conversion while preserving the configuration details.

#### input_format_configuration

* `deserializer` - (Required) Nested argument that specifies which deserializer to use. You can choose either the Apache Hive JSON SerDe or the OpenX JSON SerDe. More details below.

##### deserializer

~> **NOTE:** One of the deserializers must be configured. If no nested configuration needs to occur simply declare as `XXX_json_ser_de = []` or `XXX_json_ser_de {}`.

* `hive_json_ser_de` - (Optional) Nested argument that specifies the native Hive / HCatalog JsonSerDe. More details below.
* `open_x_json_ser_de` - (Optional) Nested argument that specifies the OpenX SerDe. More details below.

###### hive_json_ser_de

* `timestamp_formats` - (Optional) A list of how you want Kinesis Data Firehose to parse the date and time stamps that may be present in your input data JSON. To specify these format strings, follow the pattern syntax of JodaTime's DateTimeFormat format strings. For more information, see [Class DateTimeFormat](https://www.joda.org/joda-time/apidocs/org/joda/time/format/DateTimeFormat.html). You can also use the special value millis to parse time stamps in epoch milliseconds. If you don't specify a format, Kinesis Data Firehose uses java.sql.Timestamp::valueOf by default.

###### open_x_json_ser_de

* `case_insensitive` - (Optional) When set to true, which is the default, Kinesis Data Firehose converts JSON keys to lowercase before deserializing them.
* `column_to_json_key_mappings` - (Optional) A map of column names to JSON keys that aren't identical to the column names. This is useful when the JSON contains keys that are Hive keywords. For example, timestamp is a Hive keyword. If you have a JSON key named timestamp, set this parameter to `{ ts = "timestamp" }` to map this key to a column named ts.
* `convert_dots_in_json_keys_to_underscores` - (Optional) When set to `true`, specifies that the names of the keys include dots and that you want Kinesis Data Firehose to replace them with underscores. This is useful because Apache Hive does not allow dots in column names. For example, if the JSON contains a key whose name is "a.b", you can define the column name to be "a_b" when using this option. Defaults to `false`.

#### output_format_configuration

* `serializer` - (Required) Nested argument that specifies which serializer to use. You can choose either the ORC SerDe or the Parquet SerDe. More details below.

##### serializer

~> **NOTE:** One of the serializers must be configured. If no nested configuration needs to occur simply declare as `XXX_ser_de = []` or `XXX_ser_de {}`.

* `orc_ser_de` - (Optional) Nested argument that specifies converting data to the ORC format before storing it in Amazon S3. For more information, see [Apache ORC](https://orc.apache.org/docs/). More details below.
* `parquet_ser_de` - (Optional) Nested argument that specifies converting data to the Parquet format before storing it in Amazon S3. For more information, see [Apache Parquet](https://parquet.apache.org/documentation/latest/). More details below.

###### orc_ser_de

* `block_size_bytes` - (Optional) The Hadoop Distributed File System (HDFS) block size. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is 256 MiB and the minimum is 64 MiB. Kinesis Data Firehose uses this value for padding calculations.
* `bloom_filter_columns` - (Optional) A list of column names for which you want Kinesis Data Firehose to create bloom filters.
* `bloom_filter_false_positive_probability` - (Optional) The Bloom filter false positive probability (FPP). The lower the FPP, the bigger the Bloom filter. The default value is `0.05`, the minimum is `0`, and the maximum is `1`.
* `compression` - (Optional) The compression code to use over data blocks. The default is `SNAPPY`.
* `dictionary_key_threshold` - (Optional) A float that represents the fraction of the total number of non-null rows. To turn off dictionary encoding, set this fraction to a number that is less than the number of distinct keys in a dictionary. To always use dictionary encoding, set this threshold to `1`.
* `enable_padding` - (Optional) Set this to `true` to indicate that you want stripes to be padded to the HDFS block boundaries. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is `false`.
* `format_version` - (Optional) The version of the file to write. The possible values are `V0_11` and `V0_12`. The default is `V0_12`.
* `padding_tolerance` - (Optional) A float between 0 and 1 that defines the tolerance for block padding as a decimal fraction of stripe size. The default value is `0.05`, which means 5 percent of stripe size. For the default values of 64 MiB ORC stripes and 256 MiB HDFS blocks, the default block padding tolerance of 5 percent reserves a maximum of 3.2 MiB for padding within the 256 MiB block. In such a case, if the available size within the block is more than 3.2 MiB, a new, smaller stripe is inserted to fit within that space. This ensures that no stripe crosses block boundaries and causes remote reads within a node-local task. Kinesis Data Firehose ignores this parameter when `enable_padding` is `false`.
* `row_index_stride` - (Optional) The number of rows between index entries. The default is `10000` and the minimum is `1000`.
* `stripe_size_bytes` - (Optional) The number of bytes in each stripe. The default is 64 MiB and the minimum is 8 MiB.

###### parquet_ser_de

* `block_size_bytes` - (Optional) The Hadoop Distributed File System (HDFS) block size. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is 256 MiB and the minimum is 64 MiB. Kinesis Data Firehose uses this value for padding calculations.
* `compression` - (Optional) The compression code to use over data blocks. The possible values are `UNCOMPRESSED`, `SNAPPY`, and `GZIP`, with the default being `SNAPPY`. Use `SNAPPY` for higher decompression speed. Use `GZIP` if the compression ratio is more important than speed.
* `enable_dictionary_compression` - (Optional) Indicates whether to enable dictionary compression.
* `max_padding_bytes` - (Optional) The maximum amount of padding to apply. This is useful if you intend to copy the data from Amazon S3 to HDFS before querying. The default is `0`.
* `page_size_bytes` - (Optional) The Parquet page size. Column chunks are divided into pages. A page is conceptually an indivisible unit (in terms of compression and encoding). The minimum value is 64 KiB and the default is 1 MiB.
* `writer_version` - (Optional) Indicates the version of row format to output. The possible values are `V1` and `V2`. The default is `V1`.

#### schema_configuration

* `database_name` - (Required) Specifies the name of the AWS Glue database that contains the schema for the output data.
* `role_arn` - (Required) The role that Kinesis Data Firehose can use to access AWS Glue. This role must be in the same account you use for Kinesis Data Firehose. Cross-account roles aren't allowed.
* `table_name` - (Required) Specifies the AWS Glue table that contains the column information that constitutes your data schema.
* `catalog_id` - (Optional) The ID of the AWS Glue Data Catalog. If you don't supply this, the AWS account ID is used by default.
* `region` - (Optional) If you don't specify an AWS Region, the default is the current region.
* `version_id` - (Optional) Specifies the table version for the output data schema. Defaults to `LATEST`.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) specifying the Stream

[1]: https://aws.amazon.com/documentation/firehose/

## Import

Kinesis Firehose Delivery streams can be imported using the stream ARN, e.g.

```
$ terraform import aws_kinesis_firehose_delivery_stream.foo arn:aws:firehose:us-east-1:XXX:deliverystream/example
```

Note: Import does not work for stream destination `s3`. Consider using `extended_s3` since `s3` destination is deprecated.
