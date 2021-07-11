---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_flow"
description: |-
  Provides an AppFlow flow resource.
---

# Resource: aws_appflow_flow

Creates an AWS AppFlow flow.

For information about AppFlow flows, see the
[Amazon AppFlow API Reference][1]. For specific information about creating
AppFlow flow, see the [CreateFlow][2] page in the Amazon
AppFlow API Reference.

## Example Usage

```terraform
resource "aws_appflow_flow" "from_s3_to_s3" {
  flow_name   = "flow"
  description = "Flow from S3 to S3"

  destination_flow_config_list {
    connector_type = "S3"
    destination_connector_properties {
      s3 {
        bucket_name = aws_s3_bucket.destination_bucket.id
        s3_output_format_config {
          file_type = "JSON"
          aggregation_config {
            aggregation_type = "None"
          }
        }
      }
    }
  }

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket.source_bucket.id
        bucket_prefix = "emails"
      }
    }
  }

  tasks {
    source_fields = [
      "email",
    ]
    task_properties = {}
    task_type       = "Filter"
    connector_operator {
      s3 = "PROJECTION"
    }
  }
  tasks {
    destination_field = "email"
    source_fields = [
      "email",
    ]
    task_properties = {}
    task_type       = "Map"
    connector_operator {
      s3 = "NO_OP"
    }
  }

  trigger_config {
    trigger_type = "OnDemand"
  }

  depends_on = [
    aws_s3_bucket_policy.source_bucket,
    aws_s3_bucket_policy.destination_bucket,
  ]
}

resource "aws_s3_bucket" "source_bucket" {
  bucket = "source-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_policy" "source_bucket" {
  bucket = aws_s3_bucket.source_bucket.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appflow.amazonaws.com"
      },
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Resource": [
        "${aws_s3_bucket.source_bucket.arn}",
        "${aws_s3_bucket.source_bucket.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination_bucket" {
  bucket = "destination-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_policy" "destination_bucket" {
  bucket = aws_s3_bucket.destination_bucket.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appflow.amazonaws.com"
      },
      "Action": [
        "s3:PutObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts",
        "s3:ListBucketMultipartUploads",
        "s3:GetBucketAcl",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "${aws_s3_bucket.destination_bucket.arn}",
        "${aws_s3_bucket.destination_bucket.arn}/*"
      ]
    }
  ]
}
POLICY
}
```

## Argument Reference

The AppFlow flow argument layout is a complex structure.

### Top-Level Arguments

* `description` (Optional) - A description of the flow you want to create.

* `destination_flow_config_list` (Required) - The [configuration](#destination-flow-config-arguments) that controls how Amazon AppFlow places data in the destination connector.

* `flow_name` (Required) - The specified name of the flow.

* `kms_arn` (Optional) - The ARN (Amazon Resource Name) of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.

* `source_flow_config` (Required) - The [configuration](#source-flow-config-arguments) that controls how Amazon AppFlow retrieves data from the source connector.

* `tags` (Optional) - The tags used to organize, track, or control access for your flow.

* `tasks` (Required) - A list of [tasks](#task-arguments) that Amazon AppFlow performs while transferring the data in the flow run.

* `trigger_config` (Required) - The [trigger settings](#trigger-config-arguments) that determine how and when the flow runs.

#### Destination Flow Config Arguments

* `connector_profile_name` (Optional) - The name of the connector profile. This name must be unique for each connector profile in the AWS account.

* `connector_type` (Required) - The type of connector. One of: `Amplitude`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.

* `destination_connector_properties` (Required) - This stores the [information](#destination-connector-properties-arguments) that is required to query a particular connector.

##### Destination Connector Properties Arguments

* `customer_profiles` (Optional) - The [properties](#customer-profiles-destination-properties-arguments) required to query Amazon Connect Customer Profiles.

* `event_bridge` (Optional) - The [properties](#eventbridge-destination-properties-arguments) required to query Amazon EventBridge.

* `honeycode` (Optional) - The [properties](#honeycode-destination-properties-arguments) required to query Amazon Honeycode.

* `redshift` (Optional) - The [properties](#redshift-destination-properties-arguments) required to query Amazon Redshift.

* `s3` (Optional) - The [properties](#s3-destination-properties-arguments) required to query Amazon S3.

* `salesforce` (Optional) - The [properties](#salesforce-destination-properties-arguments) required to query Salesforce.

* `snowflake` (Optional) - The [properties](#snowflake-destination-properties-arguments) required to query Snowflake.

* `upsolver` (Optional) - The [properties](#upsolver-destination-properties-arguments) required to query Upsolver.

##### Customer Profiles Destination Properties Arguments

* `domain_name` (Required) - The unique name of the Amazon Connect Customer Profiles domain.

* `object_type_name` (Optional) - The object specified in the Amazon Connect Customer Profiles flow destination.

##### EventBridge Destination Properties Arguments

* `error_handling_config` (Optional) - The [settings](#error-handling-config-arguments) that determine how Amazon AppFlow handles an error when placing data in the destination. For example, this setting would determine if the flow should fail after one insertion error, or continue and attempt to insert every record regardless of the initial failure.

* `object` (Required) - The object specified in the Amazon EventBridge flow destination.

##### Honeycode Destination Properties Arguments

* `error_handling_config` (Optional) - The [settings](#error-handling-config-arguments) that determine how Amazon AppFlow handles an error when placing data in the destination. For example, this setting would determine if the flow should fail after one insertion error, or continue and attempt to insert every record regardless of the initial failure.

* `object` (Required) - The object specified in the Amazon Honeycode flow destination.

##### Redshift Destination Properties Arguments

* `bucket_prefix` (Optional) - The object key for the bucket in which Amazon AppFlow places the destination files.

* `error_handling_config` (Optional) - The [settings](#error-handling-config-arguments) that determine how Amazon AppFlow handles an error when placing data in the Amazon Redshift destination. For example, this setting would determine if the flow should fail after one insertion error, or continue and attempt to insert every record regardless of the initial failure.

* `intermediate_bucket_name` (Required) - The intermediate bucket that Amazon AppFlow uses when moving data into Amazon Redshift.

* `object` (Required) - The object specified in the Amazon Redshift flow destination.

##### S3 Destination Properties Arguments

* `bucket_name` (Required) - The Amazon S3 bucket name in which Amazon AppFlow places the transferred data.

* `bucket_prefix` (Optional) - The object key for the destination bucket in which Amazon AppFlow places the files.

* `s3_output_format_config` (Optional) - The [configuration](#s3-output-format-config-arguments) that determines how Amazon AppFlow should format the flow output data when Amazon S3 is used as the destination.

##### Salesforce Destination Properties Arguments

* `error_handling_config` (Optional) - The [settings](#error-handling-config-arguments) that determine how Amazon AppFlow handles an error when placing data in the Salesforce destination. For example, this setting would determine if the flow should fail after one insertion error, or continue and attempt to insert every record regardless of the initial failure. ErrorHandlingConfig is a part of the destination connector details.

* `id_field_names` (Optional) - The name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete. An array of 0 or 1 string(s).

* `object` (Required) - The object specified in the Salesforce flow destination.

* `write_operation_type` (Optional) - This specifies the type of write operation to be performed in Salesforce. When the value is `UPSERT`, then `id_field_names` is required. Valid values: `INSERT`, `UPSERT`, `UPDATE`.

##### Snowflake Destination Properties Arguments

* `bucket_prefix` (Optional) - The object key for the destination bucket in which Amazon AppFlow places the files.

* `error_handling_config` (Optional) - The [settings](#error-handling-config-arguments) that determine how Amazon AppFlow handles an error when placing data in the Snowflake destination. For example, this setting would determine if the flow should fail after one insertion error, or continue and attempt to insert every record regardless of the initial failure.

* `intermediate_bucket_name` (Required) - The intermediate bucket that Amazon AppFlow uses when moving data into Snowflake.

* `object` (Required) - The object specified in the Snowflake flow destination.

##### Upsolver Destination Properties Arguments

* `bucket_name` (Required) - The Upsolver Amazon S3 bucket name in which Amazon AppFlow places the transferred data. Name has to start with `upsolver-appflow`.

* `bucket_prefix` (Optional) - The object key for the destination Upsolver Amazon S3 bucket in which Amazon AppFlow places the files.

* `s3_output_format_config` (Optional) - The [configuration](#upsolver-s3-output-format-config-arguments) that determines how data is formatted when Upsolver is used as the flow destination.

##### Error Handling Config Arguments

* `bucket_name` (Optional) - Specifies the name of the Amazon S3 bucket.

* `bucket_prefix` (Optional) - Specifies the Amazon S3 bucket prefix.

* `fail_on_first_destination_error` (Optional) - Specifies if the flow should fail after the first instance of a failure when attempting to place data in the destination. Valid values: `false`, `true`.

##### S3 Output Format Config Arguments

* `aggregation_config` (Optional) - The [aggregation settings](#aggregation-config-arguments) that you can use to customize the output format of your flow data.

* `file_type` (Optional) - Indicates the file type that Amazon AppFlow places in the Amazon S3 bucket. Valid values: `CSV`, `JSON`, `PARQUET`.

* `prefix_config` (Optional) - The [configuration](#prefix-config-arguments) that determines the prefix that Amazon AppFlow applies to the folder name in the Amazon S3 bucket. You can name folders according to the flow frequency and date.

##### Upsolver S3 Output Format Config Arguments

* `aggregation_config` (Optional) - The [aggregation settings](#aggregation-config-arguments) that you can use to customize the output format of your flow data.

* `file_type` (Optional) - Indicates the file type that Amazon AppFlow places in the Upsolver Amazon S3 bucket. Valid values: `CSV`, `JSON`, `PARQUET`.

* `prefix_config` (Required) - The [configuration](#prefix-config-arguments) that determines the prefix that Amazon AppFlow applies to the destination folder name. You can name your destination folders according to the flow frequency and date.

##### Aggregation Config Arguments

* `aggregation_type` (Optional) - Specifies whether Amazon AppFlow aggregates the flow records into a single file, or leave them unaggregated. Valid values: `None`, `SingleFile`.

##### Prefix Config Arguments

* `prefix_format` (Optional) - Determines the level of granularity that's included in the prefix. Valid values: `YEAR`, `MONTH`, `DAY`, `HOUR`, `MINUTE`.

* `prefix_type` (Optional) - Determines the format of the prefix, and whether it applies to the file name, file path, or both. Valid values: `FILENAME`, `PATH`, `PATH_AND_FILENAME`.

#### Source Flow Config Arguments

* `connector_profile_name` (Optional) - The name of the connector profile. This name must be unique for each connector profile in the AWS account.

* `connector_type` (Required) - The type of connector. One of: `Amplitude`, `CustomerProfiles`, `Datadog`, `Dynatrace`, `EventBridge`, `Googleanalytics`, `Honeycode`, `Infornexus`, `LookoutMetrics`, `Marketo`, `Redshift`, `S3`, `Salesforce`, `Servicenow`, `Singular`, `Slack`, `Snowflake`, `Trendmicro`, `Upsolver`, `Veeva`, `Zendesk`.

* `incremental_pull_config` (Optional) - Defines the [configuration](#incremental-pull-config-arguments) for a scheduled incremental data pull. If a valid configuration is provided, the fields specified in the configuration are used when querying for the incremental data pull.

* `source_connector_properties` (Required) - The [configuration](#source-connector-properties-arguments) that specifies the information that is required to query a particular source connector.

##### Incremental Pull Config Arguments

* `datetime_type_field_name` (Optional) - A field that specifies the date time or timestamp field as the criteria to use when importing incremental records from the source.

#### Source Connector Properties Arguments

* `amplitude` (Optional) - Specifies the [information](#amplitude-source-properties-arguments) that is required for querying Amplitude.

* `datadog` (Optional) - Specifies the [information](#datadog-source-properties-arguments) that is required for querying Datadog.

* `dynatrace` (Optional) - Specifies the [information](#dynatrace-source-properties-arguments) that is required for querying Dynatrace.

* `google_analytics` (Optional) - Specifies the [information](#google-analytics-source-properties-arguments) that is required for querying Google Analytics.

* `infor_nexus` (Optional) - Specifies the [information](#infor-nexus-source-properties-arguments) that is required for querying Infor Nexus.

* `marketo` (Optional) - Specifies the [information](#marketo-source-properties-arguments) that is required for querying Marketo.

* `s3` (Optional) - Specifies the [information](#s3-source-properties-arguments) that is required for querying Amazon S3.

* `salesforce` (Optional) - Specifies the [information](#salesforce-source-properties-arguments) that is required for querying Salesforce.

* `service_now` (Optional) - Specifies the [information](#servicenow-source-properties-arguments) that is required for querying ServiceNow.

* `singular` (Optional) - Specifies the [information](#singular-source-properties-arguments) that is required for querying Singular.

* `slack` (Optional) - Specifies the [information](#slack-source-properties-arguments) that is required for querying Slack.

* `trendmicro` (Optional) - Specifies the [information](#trend-micro-source-properties-arguments) that is required for querying Trend Micro.

* `veeva` (Optional) - Specifies the [information](#veeva-source-properties-arguments) that is required for querying Veeva.

* `zendesk` (Optional) - Specifies the [information](#zendesk-source-properties-arguments) that is required for querying Zendesk.

##### Amplitude Source Properties Arguments

* `object` (Required) - The object specified in the Amplitude flow source.

##### Datadog Source Properties Arguments

* `object` (Required) - The object specified in the Datadog flow source.

##### Dynatrace Source Properties Arguments

* `object` (Required) - The object specified in the Dynatrace flow source.

##### Google Analytics Source Properties Arguments

* `object` (Required) - The object specified in the Google Analytics flow source.

##### Infor Nexus Source Properties Arguments

* `object` (Required) - The object specified in the Infor Nexus flow source.

##### Marketo Source Properties Arguments

* `object` (Required) - The object specified in the Marketo flow source.

##### S3 Source Properties Arguments

* `bucket_name` (Required) - The Amazon S3 bucket name where the source files are stored.

* `bucket_prefix` (Optional) - The object key for the Amazon S3 bucket in which the source files are stored.

##### Salesforce Source Properties Arguments

* `enable_dynamic_field_update` (Optional) - The flag that enables dynamic fetching of new (recently added) fields in the Salesforce objects while running a flow. Valid values: `false`, `true`.

* `include_deleted_records` (Optional) - Indicates whether Amazon AppFlow includes deleted files in the flow run. Valid values: `false`, `true`.

* `object` (Required) - The object specified in the Salesforce flow source.

##### ServiceNow Source Properties Arguments

* `object` (Required) - The object specified in the ServiceNow flow source.

##### Singular Source Properties Arguments

* `object` (Required) - The object specified in the Singular flow source.

##### Slack Source Properties Arguments

* `object` (Required) - The object specified in the Slack flow source.

##### Trend Micro Source Properties Arguments

* `object` (Required) - The object specified in the Trend Micro flow source.

##### Veeva Source Properties Arguments

* `object` (Required) - The object specified in the Veeva flow source.

##### Zendesk Source Properties Arguments

* `object` (Required) - The object specified in the Zendesk flow source.

#### Task Arguments

* `connector_operator` (Optional) - The [operation](#connector-operator-arguments) to be performed on the provided source fields.

* `destination_field` (Optional) - A field in a destination connector, or a field value against which Amazon AppFlow validates a source field.

* `source_fields` (Required) - The source fields to which a particular task is applied. A list of strings.

* `task_properties` (Required) - A map used to store task-related information. The execution service looks for particular information based on the `task_type`. Valid keys: `VALUE`, `VALUES`, `DATA_TYPE`, `UPPER_BOUND`, `LOWER_BOUND`, `SOURCE_DATA_TYPE`, `DESTINATION_DATA_TYPE`, `VALIDATION_ACTION`, `MASK_VALUE`, `MASK_LENGTH`, `TRUNCATE_LENGTH`, `MATH_OPERATION_FIELDS_ORDER`, `CONCAT_FORMAT`, `SUBFIELD_CATEGORY_MAP`. A map can be empty.

* `task_type` (Required) - Specifies the particular task implementation that Amazon AppFlow performs. Valid values: `Arithmetic`, `Filter`, `Map`, `Mask`, `Merge`, `Truncate`, `Validate`.

##### Connector Operator Arguments

* `amplitude` (Optional) - The operation to be performed on the provided Amplitude source fields. Valid values: `BETWEEN`

* `datadog` (Optional) - The operation to be performed on the provided Datadog source fields. Valid values: `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `dynatrace` (Optional) - The operation to be performed on the provided Dynatrace source fields. Valid values: `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `google_analytics` (Optional) - The operation to be performed on the provided Google Analytics source fields. Valid values: `PROJECTION`, `BETWEEN`

* `infor_nexus` (Optional) - The operation to be performed on the provided Infor Nexus source fields. Valid values: `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `marketo` (Optional) - The operation to be performed on the provided Marketo source fields. Valid values: `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `s3` (Optional) - The operation to be performed on the provided Amazon S3 source fields. Valid values: `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `salesforce` (Optional) - The operation to be performed on the provided Salesforce source fields. Valid values: `PROJECTION`, `LESS_THAN`, `CONTAINS`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `service_now` (Optional) - The operation to be performed on the provided ServiceNow source fields. Valid values: `PROJECTION`, `CONTAINS`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `singular` (Optional) - The operation to be performed on the provided Singular source fields. Valid values: `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `slack` (Optional) - The operation to be performed on the provided Slack source fields. Valid values: `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `trendmicro` (Optional) - The operation to be performed on the provided Trend Micro source fields. Valid values: `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `veeva` (Optional) - The operation to be performed on the provided Veeva source fields. Valid values: `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

* `zendesk` (Optional) - The operation to be performed on the provided Zendesk source fields. Valid values: `PROJECTION`, `GREATER_THAN`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, `NO_OP`

#### Trigger Config Arguments

* `trigger_properties` (Optional) - Specifies the [configuration details](#trigger-properties-arguments) of a schedule-triggered flow as defined by the user. Currently, these settings only apply to the `Scheduled` trigger type.

* `trigger_type` (Required) - Specifies the type of flow trigger. This can be `OnDemand`, `Scheduled`, or `Event`.

##### Trigger Properties Arguments

* `scheduled` (Optional) - Specifies the [configuration details](#scheduled-trigger-properties-arguments) of a schedule-triggered flow as defined by the user.

##### Scheduled Trigger Properties Arguments

* `data_pull_mode` (Optional) - Specifies whether a scheduled flow has an incremental data transfer or a complete data transfer for each flow run. Valid values: `Incremental`, `Complete`.

* `first_execution_from` (Optional) - Specifies the date range for the records to import from the connector in the first flow run. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

* `schedule_end_time` (Optional) - Specifies the scheduled end time for a schedule-triggered flow. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

* `schedule_expression` (Required) - The scheduling expression that determines the rate at which the schedule will run, for example `rate(5minutes)`. For more information, see [Schedule expressions for rules][3] in the CloudWatch Events User Guide.

* `schedule_offset` (Optional) - Specifies the optional offset that is added to the time interval for a schedule-triggered flow. Between 0 and 36000.

* `schedule_start_time` (Optional) - Specifies the scheduled start time for a schedule-triggered flow.Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

* `timezone` (Optional) - Specifies the time zone used when referring to the date and time of a scheduled-triggered flow, such as `America/New_York`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_at` - Specifies when the flow was created.

* `created_by` - The ARN of the user who created the flow.

* `flow_arn` - The flow's Amazon Resource Name (ARN).

* `flow_status` - Indicates the current status of the flow.

## Import

AppFlow flows can be imported using the flow name, e.g.

```
$ terraform import aws_appflow_flow.flow flow-name
```

[1]: https://docs.aws.amazon.com/appflow/1.0/APIReference/Welcome.html
[2]: https://docs.aws.amazon.com/appflow/1.0/APIReference/API_CreateFlow.html
[3]: https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html
