---
layout: "aws"
page_title: "AWS: aws_kinesis_analytics_application"
sidebar_current: "docs-aws-resource-kinesis-analytics-application"
description: |-
  Provides a AWS Kinesis Analytics Application
---

# aws_kinesis_analytics_application

Provides a Kinesis Analytics Application resource. Kinesis Analytics is a managed service that
allows processing and analyzing streaming data using standard SQL.

For more details, see the [Amazon Kinesis Analytics Documentation][1].

## Example Usage

```hcl
resource "aws_kinesis_stream" "test_stream" {
  name        = "terraform-kinesis-test"
  shard_count = 1
}

resource "aws_kinesis_analytics_application" "test_application" {
  name = "kinesis-analytics-application-test"

  inputs {
    name_prefix = "test_prefix"
    kinesis_stream {
      resource_arn = "${aws_kinesis_stream.test_stream.arn}"
      role_arn     = "${aws_iam_role.test.arn}"
    }
    parallelism {
      count = 1
    }
    schema {
      record_columns {
        mapping  = "$.test"
        name     = "test"
        sql_type = "VARCHAR(8)"
      }
      record_encoding = "UTF-8"
      record_format {
        mapping_parameters {
          json {
            record_row_path = "$"
          }
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Kinesis Analytics Application.
* `code` - (Optional) SQL Code to transform input data, and generate output.
* `description` - (Optional) Description of the application.
* `cloudwatch_logging_options` - (Optional) The CloudWatch log stream options to monitor application errors. 
See [CloudWatch Logging Options](#cloudwatch-logging-options) below for more details.
* `inputs` - (Optional) Input configuration of the application. See [Inputs](#inputs) below for more details.
* `outputs` - (Optional) Output destination configuration of the application. See [Outputs](#outputs) below for more details.
* `reference_data_sources` - (Optional) An S3 Reference Data Source for the application. 
See [Reference Data Sources](#reference-data-sources) below for more details.

### CloudWatch Logging Options

Configure a CloudWatch Log Stream to monitor application errors.

The `cloudwatch_logging_options` block supports the following:

* `log_stream_arn` - (Required) The ARN of the CloudWatch Log Stream.
* `role_arn` - (Required) The ARN of the IAM Role used to send application messages.

### Inputs

Configure an Input for the Kinesis Analytics Application. You can only have 1 Input configured.

The `inputs` block supports the following:

* `name_prefix` - (Required) The Name Prefix to use when creating an in-application stream.
* `schema` - (Required) The Schema format of the data in the streaming source. See [Source Schema](#source-schema) below for more details.
* `kinesis_firehose` - (Optional) The Kinesis Firehose configuration for the streaming source. Conflicts with `kinesis_stream`. 
See [Kinesis Firehose](#kinesis-firehose) below for more details.
* `kinesis_stream` - (Optional) The Kinesis Stream configuration for the streaming source. Conflicts with `kinesis_firehose`. 
See [Kinesis Stream](#kinesis-stream) below for more details.
* `parallelism` - (Optional) The number of Parallel in-application streams to create. 
See [Parallelism](#parallelism) below for more details.
* `processing_configuration` - (Optional) The Processing Configuration to transform records as they are received from the stream.
See [Processing Configuration](#processing-configuration) below for more details.

### Outputs

Configure Output destinations for the Kinesis Analytics Application. You can have a maximum of 3 destinations configured.

The `outputs` block supports the following:

* `name` - (Required) The Name of the in-application stream.
* `schema` - (Required) The Schema format of the data written to the destination. See [Destination Schema](#destination-schema) below for more details.
* `kinesis_firehose` - (Optional) The Kinesis Firehose configuration for the destination stream. Conflicts with `kinesis_stream`.
See [Kinesis Firehose](#kinesis-firehose) below for more details.
* `kinesis_stream` - (Optional) The Kinesis Stream configuration for the destination stream. Conflicts with `kinesis_firehose`.
See [Kinesis Stream](#kinesis-stream) below for more details.
* `lambda` - (Optional) The Lambda function destination. See [Lambda](#lambda) below for more details.

### Reference Data Sources

Add a Reference Data Source to the Kinesis Analytics Application. You can only have 1 Reference Data Source.

The `reference_data_sources` block supports the following:

* `schema` - (Required) The Schema format of the data in the streaming source. See [Source Schema](#source-schema) below for more details.
* `table_name` - (Required) The in-application Table Name.
* `s3` - (Optional) The S3 configuration for the reference data source. See [S3 Reference](#s3-reference) below for more details. 

#### Kinesis Firehose

Configuration for a Kinesis Firehose delivery stream.

The `kinesis_firehose` block supports the following:

* `resource_arn` - (Required) The ARN of the Kinesis Firehose delivery stream.
* `role_arn` - (Required) The ARN of the IAM Role used to access the stream.

#### Kinesis Stream

Configuration for a Kinesis Stream.

The `kinesis_stream` block supports the following:

* `resource_arn` - (Required) The ARN of the Kinesis Stream.
* `role_arn` - (Required) The ARN of the IAM Role used to access the stream.

#### Destination Schema

The Schema format of the data in the destination.

The `schema` block supports the following:

* `record_format_type` - (Required) The Format Type of the records on the output stream. Can be `CSV` or `JSON`.

#### Source Schema

The Schema format of the data in the streaming source.

The `schema` block supports the following:

* `record_columns` - (Required) The Record Column mapping for the streaming source data element.
See [Record Columns](#record-columns) below for more details.
* `record_format` - (Required) The Record Format and mapping information to schematize a record.
See [Record Format](#record-format) below for more details.
* `record_encoding` - (Optional) The Encoding of the record in the streaming source.

#### Parallelism

Configures the number of Parallel in-application streams to create.

The `parallelism` block supports the following:

* `count` - (Required) The Count of streams.

#### Processing Configuration

The Processing Configuration to transform records as they are received from the stream.

The `processing_configuration` block supports the following:

* `lambda` - (Required) The Lambda function configuration. See [Lambda](#lambda) below for more details.

#### Lambda

The Lambda function that pre-processes records in the stream.

The `lambda` block supports the following:

* `resource_arn` - (Required) The ARN of the Lambda function.
* `role_arn` - (Required) The ARN of the IAM Role used to access the Lambda function.

#### Record Columns

The Column mapping of each data element in the streaming source to the corresponding column in the in-application stream.

The `record_columns` block supports the following:

* `name` - (Required) Name of the column.
* `sql_type` - (Required) The SQL Type of the column.
* `mapping` - (Optional) The Mapping reference to the data element.

#### Record Format

The Record Format and relevant mapping information that should be applied to schematize the records on the stream.

The `record_format` block supports the following:

* `record_format_type` - (Required) The type of Record Format. Can be `CSV` or `JSON`.
* `mapping_parameters` - (Optional) The Mapping Information for the record format.
See [Mapping Parameters](#mapping-parameters) below for more details.

#### Mapping Parameters

Provides Mapping information specific to the record format on the streaming source.

The `mapping_parameters` block supports the following:

* `csv` - (Optional) Mapping information when the record format uses delimiters.
See [CSV Mapping Parameters](#csv-mapping-parameters) below for more details.
* `json` - (Optional) Mapping information when JSON is the record format on the streaming source.
See [JSON Mapping Parameters](#json-mapping-parameters) below for more details.

#### CSV Mapping Parameters

Mapping information when the record format uses delimiters.

The `csv` block supports the following:

* `record_column_delimiter` - (Required) The Column Delimiter.
* `record_row_delimiter` - (Required) The Row Delimiter.

#### JSON Mapping Parameters

Mapping information when JSON is the record format on the streaming source.

The `json` block supports the following:

* `record_row_path` - (Required) Path to the top-level parent that contains the records.

#### S3 Reference

Identifies the S3 bucket and object that contains the reference data.

The `s3` blcok supports the following:

* `bucket_arn` - (Required) The S3 Bucket ARN.
* `file_key` - (Required) The File Key name containing reference data.
* `role_arn` - (Required) The IAM Role ARN to read the data.

## Attributes Reference

The following attributes are exported along with all argument references:

* `id` - The ARN of the Kinesis Analytics Application.
* `arn` - The ARN of the Kinesis Analytics Appliation.
* `create_timestamp` - The Timestamp when the application version was created.
* `last_update_timestamp` - The Timestamp when the application was last updated.
* `status` - The Status of the application.
* `version` - The Version of the application.

[1]: https://docs.aws.amazon.com/kinesisanalytics/latest/dev/what-is.html
