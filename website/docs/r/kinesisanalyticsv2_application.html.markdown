---
subcategory: "Kinesis Data Analytics v2 (SQL and Flink Applications)"
layout: "aws"
page_title: "AWS: aws_kinesisanalyticsv2_application"
description: |-
  Manages a Kinesis Analytics v2 Application.
---

# Resource: aws_kinesisanalyticsv2_application

Manages a Kinesis Analytics v2 Application.
This resource can be used to manage both Kinesis Data Analytics for SQL applications and Kinesis Data Analytics for Apache Flink applications.

-> **Note:** Kinesis Data Analytics for SQL applications created using this resource cannot currently be viewed in the AWS Console. To manage Kinesis Data Analytics for SQL applications that can also be viewed in the AWS Console, use the [`aws_kinesis_analytics_application`](/docs/providers/aws/r/kinesis_analytics_application.html) resource.

## Example Usage

### Apache Flink Application

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example-flink-application"
}

resource "aws_s3_bucket_object" "example" {
  bucket = aws_s3_bucket.example.bucket
  key    = "example-flink-application"
  source = "flink-app.jar"
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = aws_s3_bucket_object.example.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-1"

        property_map = {
          Key1 = "Value1"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-2"

        property_map = {
          KeyA = "ValueA"
          KeyB = "ValueB"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }

  tags = {
    Environment = "test"
  }
}
```

### SQL Application

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name = "example-sql-application"
}

resource "aws_cloudwatch_log_stream" "example" {
  name           = "example-sql-application"
  log_group_name = aws_cloudwatch_log_group.example.name
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-sql-application"
  runtime_environment    = "SQL-1.0"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "PREFIX_1"

        input_parallelism {
          count = 3
        }

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-1"
          }

          record_column {
            name     = "COLUMN_2"
            sql_type = "DOUBLE"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "\n"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.example.arn
        }
      }

      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "JSON"
        }

        lambda_output {
          resource_arn = aws_lambda_function.example.arn
        }
      }

      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.example.arn
        }
      }

      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = "KEY-1"
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.example.arn
  }
}
```

### VPC Configuration

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example-flink-application"
}

resource "aws_s3_bucket_object" "example" {
  bucket = aws_s3_bucket.example.bucket
  key    = "example-flink-application"
  source = "flink-app.jar"
}

resource "aws_kinesisanalyticsv2_application" "example" {
  name                   = "example-flink-application"
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.example.arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.example.arn
          file_key   = aws_s3_bucket_object.example.key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = [aws_security_group.example[0].id, aws_security_group.example[1].id]
      subnet_ids         = [aws_subnet.example.id]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application.
* `runtime_environment` - (Required) The runtime environment for the application. Valid values: `SQL-1_0`, `FLINK-1_6`, `FLINK-1_8`, `FLINK-1_11`.
* `service_execution_role` - (Required) The ARN of the [IAM role](/docs/providers/aws/r/iam_role.html) used by the application to access Kinesis data streams, Kinesis Data Firehose delivery streams, Amazon S3 objects, and other external resources.
* `application_configuration` - (Optional) The application's configuration
* `cloudwatch_logging_options` - (Optional) A [CloudWatch log stream](/docs/providers/aws/r/cloudwatch_log_stream.html) to monitor application configuration errors.
* `description` - (Optional) A summary description of the application.
* `tags` - (Optional) A map of tags to assign to the application.

The `application_configuration` object supports the following:

* `application_code_configuration` - (Required) The code location and type parameters for the application.
* `application_snapshot_configuration` - (Optional) Describes whether snapshots are enabled for a Flink-based application.
* `environment_properties` - (Optional) Describes execution properties for a Flink-based application.
* `flink_application_configuration` - (Optional) The configuration of a Flink-based application.
* `sql_application_configuration` - (Optional) The configuration of a SQL-based application.
* `vpc_configuration` - (Optional) The VPC configuration of a Flink-based application.

The `application_code_configuration` object supports the following:

* `code_content_type` - (Required) Specifies whether the code content is in text or zip format. Valid values: `PLAINTEXT`, `ZIPFILE`.
* `code_content` - (Optional) The location and type of the application code.

The `code_content` object supports the following:

* `s3_content_location` - (Optional) Information about the Amazon S3 bucket containing the application code.
* `text_content` - (Optional) The text-format code for the application.

The `s3_content_location` object supports the following:

* `bucket_arn` - (Required) The ARN for the S3 bucket containing the application code.
* `file_key` - (Required) The file key for the object containing the application code.
* `object_version` - (Optional) The version of the object containing the application code.

The `application_snapshot_configuration` object supports the following:

* `snapshots_enabled` - (Required) Describes whether snapshots are enabled for a Flink-based Kinesis Data Analytics application.

The `environment_properties` object supports the following:

* `property_group` - (Required) Describes the execution property groups.

The `property_group` object supports the following:

* `property_group_id` - (Required) The key of the application execution property key-value map.
* `property_map` - (Required) Application execution property key-value map.

The `flink_application_configuration` object supports the following:

* `checkpoint_configuration` - (Optional) Describes an application's checkpointing configuration.
* `monitoring_configuration` - (Optional) Describes configuration parameters for CloudWatch logging for an application.
* `parallelism_configuration` - (Optional) Describes parameters for how an application executes multiple tasks simultaneously.

The `checkpoint_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether the application uses Kinesis Data Analytics' default checkpointing behavior. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `checkpointing_enabled`, `checkpoint_interval`, or `min_pause_between_checkpoints` attribute values to be effective. If this attribute is set to `DEFAULT`, the application will always use the following values:
    * `checkpointing_enabled = true`
    * `checkpoint_interval = 60000`
    * `min_pause_between_checkpoints = 5000`
* `checkpointing_enabled` - (Optional) Describes whether checkpointing is enabled for a Flink-based Kinesis Data Analytics application.
* `checkpoint_interval` - (Optional) Describes the interval in milliseconds between checkpoint operations.
* `min_pause_between_checkpoints` - (Optional) Describes the minimum time in milliseconds after a checkpoint operation completes that a new checkpoint operation can start.

The `monitoring_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether to use the default CloudWatch logging configuration for an application. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `log_level` or `metrics_level` attribute values to be effective.
* `log_level` - (Optional) Describes the verbosity of the CloudWatch Logs for an application. Valid values: `DEBUG`, `ERROR`, `INFO`, `WARN`.
* `metrics_level` - (Optional) Describes the granularity of the CloudWatch Logs for an application. Valid values: `APPLICATION`, `OPERATOR`, `PARALLELISM`, `TASK`.

The `parallelism_configuration` object supports the following:

* `configuration_type` - (Required) Describes whether the application uses the default parallelism for the Kinesis Data Analytics service. Valid values: `CUSTOM`, `DEFAULT`. Set this attribute to `CUSTOM` in order for any specified `auto_scaling_enabled`, `parallelism`, or `parallelism_per_kpu` attribute values to be effective.
* `auto_scaling_enabled` - (Optional) Describes whether the Kinesis Data Analytics service can increase the parallelism of the application in response to increased throughput.
* `parallelism` - (Optional) Describes the initial number of parallel tasks that a Flink-based Kinesis Data Analytics application can perform.
* `parallelism_per_kpu` - (Optional) Describes the number of parallel tasks that a Flink-based Kinesis Data Analytics application can perform per Kinesis Processing Unit (KPU) used by the application.

The `sql_application_configuration` object supports the following:

* `input` - (Optional) The input stream used by the application.
* `output` - (Optional) The destination streams used by the application.
* `reference_data_source` - (Optional) The reference data source used by the application.

The `input` object supports the following:

* `input_schema` - (Required) Describes the format of the data in the streaming source, and how each data element maps to corresponding columns in the in-application stream that is being created.
* `name_prefix` - (Required) The name prefix to use when creating an in-application stream.
* `input_parallelism` - (Optional) Describes the number of in-application streams to create.
* `input_processing_configuration` - (Optional) The input processing configuration for the input.
An input processor transforms records as they are received from the stream, before the application's SQL code executes.
* `kinesis_firehose_input` - (Optional) If the streaming source is a [Kinesis Data Firehose delivery stream](/docs/providers/aws/r/kinesis_firehose_delivery_stream.html), identifies the delivery stream's ARN.
* `kinesis_streams_input` - (Optional) If the streaming source is a [Kinesis data stream](/docs/providers/aws/r/kinesis_stream.html), identifies the stream's Amazon Resource Name (ARN).

The `input_parallelism` object supports the following:

* `count` - (Optional) The number of in-application streams to create.

The `input_processing_configuration` object supports the following:

* `input_lambda_processor` - (Required) Describes the [Lambda function](/docs/providers/aws/r/lambda_function.html) that is used to preprocess the records in the stream before being processed by your application code.

The `input_lambda_processor` object supports the following:

* `resource_arn` - (Required) The ARN of the Lambda function that operates on records in the stream.

The `input_schema` object supports the following:

* `record_column` - (Required) Describes the mapping of each data element in the streaming source to the corresponding column in the in-application stream.
* `record_format` - (Required) Specifies the format of the records on the streaming source.
* `record_encoding` - (Optional) Specifies the encoding of the records in the streaming source. For example, `UTF-8`.

The `record_column` object supports the following:

* `name` - (Required) The name of the column that is created in the in-application input stream or reference table.
* `sql_type` - (Required) The type of column created in the in-application input stream or reference table.
* `mapping` - (Optional) A reference to the data element in the streaming input or the reference data source.

The `record_format` object supports the following:

* `mapping_parameters` - (Required) Provides additional mapping information specific to the record format (such as JSON, CSV, or record fields delimited by some delimiter) on the streaming source.
* `record_format_type` - (Required) The type of record format. Valid values: `CSV`, `JSON`.

The `mapping_parameters` object supports the following:

* `csv_mapping_parameters` - (Optional) Provides additional mapping information when the record format uses delimiters (for example, CSV).
* `json_mapping_parameters` - (Optional) Provides additional mapping information when JSON is the record format on the streaming source.

The `csv_mapping_parameters` object supports the following:

* `record_column_delimiter` - (Required) The column delimiter. For example, in a CSV format, a comma (`,`) is the typical column delimiter.
* `record_row_delimiter` - (Required) The row delimiter. For example, in a CSV format, `\n` is the typical row delimiter.

The `json_mapping_parameters` object supports the following:

* `record_row_path` - (Required) The path to the top-level parent that contains the records.

The `kinesis_firehose_input` object supports the following:

* `resource_arn` - (Required) The ARN of the delivery stream.

The `kinesis_streams_input` object supports the following:

* `resource_arn` - (Required) The ARN of the input Kinesis data stream to read.

The `output` object supports the following:

* `destination_schema` - (Required) Describes the data format when records are written to the destination.
* `name` - (Required) The name of the in-application stream.
* `kinesis_firehose_output` - (Optional) Identifies a [Kinesis Data Firehose delivery stream](/docs/providers/aws/r/kinesis_firehose_delivery_stream.html) as the destination.
* `kinesis_streams_output` - (Optional) Identifies a [Kinesis data stream](/docs/providers/aws/r/kinesis_stream.html) as the destination.
* `lambda_output` - (Optional) Identifies a [Lambda function](/docs/providers/aws/r/lambda_function.html) as the destination.

The `destination_schema` object supports the following:

* `record_format_type` - (Required) Specifies the format of the records on the output stream. Valid values: `CSV`, `JSON`.

The `kinesis_firehose_output` object supports the following:

* `resource_arn` - (Required) The ARN of the destination delivery stream to write to.

The `kinesis_streams_output` object supports the following:

* `resource_arn` - (Required) The ARN of the destination Kinesis data stream to write to.

The `lambda_output` object supports the following:

* `resource_arn` - (Required) The ARN of the destination Lambda function to write to.

The `reference_data_source` object supports the following:

* `reference_schema` - (Required) Describes the format of the data in the streaming source, and how each data element maps to corresponding columns created in the in-application stream.
* `s3_reference_data_source` - (Required) Identifies the S3 bucket and object that contains the reference data.
* `table_name` - (Required) The name of the in-application table to create.

The `reference_schema` object supports the following:

* `record_column` - (Required) Describes the mapping of each data element in the streaming source to the corresponding column in the in-application stream.
* `record_format` - (Required) Specifies the format of the records on the streaming source.
* `record_encoding` - (Optional) Specifies the encoding of the records in the streaming source. For example, `UTF-8`.

The `s3_reference_data_source` object supports the following:

* `bucket_arn` - (Required) The ARN of the S3 bucket.
* `file_key` - (Required) The object key name containing the reference data.

The `vpc_configuration` object supports the following:

* `security_group_ids` - (Required) The [Security Group](/docs/providers/aws/r/security_group.html) IDs used by the VPC configuration.
* `subnet_ids` - (Required) The [Subnet](/docs/providers/aws/r/subnet.html) IDs used by the VPC configuration.

The `cloudwatch_logging_options` object supports the following:

* `log_stream_arn` - (Required) The ARN of the CloudWatch log stream to receive application messages.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The application identifier.
* `arn` - The ARN of the application.
* `create_timestamp` - The current timestamp when the application was created.
* `last_update_timestamp` - The current timestamp when the application was last updated.
* `status` - The status of the application.
* `version_id` - The current application version. Kinesis Data Analytics updates the `version_id` each time the application is updated.

## Import

`aws_kinesisanalyticsv2_application` can be imported by using the application ARN, e.g.

```
$ terraform import aws_kinesisanalyticsv2_application.example arn:aws:kinesisanalytics:us-west-2:123456789012:application/example-sql-application
```
