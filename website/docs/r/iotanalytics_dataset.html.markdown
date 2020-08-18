---
subcategory: "IoTAnalytics"
layout: "aws"
page_title: "AWS: aws_iotevents_detector_model"
sidebar_current: "docs-aws-resource-iotevents-detector-model"
description: |-
    Creates and manages an AWS IoTEvents detector model
---

# Resource: aws_iotevents_detector_model

## Example Usage

```hcl
resource "aws_iotanalytics_dataset" "dataset" {
  name = "dataset"

  action {
	  name = "action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}
		  sql_query = "select * from datastore"
	  }
  }

}
```

## Argument Reference

* `name` - (Required) The name of the data set.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the dataset.
* [`action`](#action) - Object (At least 1). A list of actions that create the data set contents.
* [`content_delivery_rule`](#content_delivery_rule) - Object (Optional) Can be many. When data set contents are created they are delivered to destinations specified here.
* [`trigger`](#trigger) - Object (Optional) Can be many. A list of triggers. A trigger causes data set contents to be populated at a specified time interval or when another data set's contents are created. The list of triggers can be empty or contain up to five DataSetTrigger objects.
* [`retention_period`](#retention_period) - Object (Optional) Only one. How long, in days, versions of data set contents are kept for the data set. If not specified or set to null, versions of data set contents are retained for at most 90 days. The number of versions of data set contents retained is determined by the versioningConfiguration parameter. (For more information, see https://docs.aws.amazon.com/iotanalytics/latest/userguide/getting-started.html#aws-iot-analytics-dataset-versions)
* [`versioning_configuration`](#versioning_configuration) - Object (Optional) Only one. How many versions of data set contents are kept. If not specified or set to null, only the latest version plus the latest succeeded version (if they are different) are kept for the time period specified by the "retentionPeriod" parameter. (For more information, see https://docs.aws.amazon.com/iotanalytics/latest/userguide/getting-started.html#aws-iot-analytics-dataset-versions)


<a name="action"><a/> The `action` argument reference.
* `name` - (Required) The name of the data set action by which data set contents are automatically created.
* [`container_action`](#container_action) - <b>Doest Not Supported Yet<b/> Object (Optional) Only one. Information which allows the system to run a containerized application in order to create the data set contents. The application must be in a Docker container along with any needed support libraries.
* [`query_action`](#query_action) - Object (Optional) Only one. An "SqlQueryDatasetAction" object that uses an SQL query to automatically create data set contents.

<a name="container_action"><a/> The `container_action` argument reference.
* `image` - (Required). The ARN of the Docker container stored in your account. The Docker container contains an application and needed support libraries and is used to generate data set contents.
* `execution_role_arn` - (Required). The ARN of the role which gives permission to the system to access needed resources in order to run the "containerAction". This includes, at minimum, permission to retrieve the data set contents which are the input to the containerized application.
* [`resource_configuration`](#resource_configuration) - Object (Optional) Only one. Configuration of the resource which executes the "containerAction".
* [`variable`](#variable) - Object (Optional) Can be many. The values of variables used within the context of the execution of the containerized application (basically, parameters passed to the application). Each variable must have a name and a value given by one of "string_value", "dataset_content_version_value", or "output_file_uri_value".

<a name="resource_configuration"><a/> The `resource_configuration` argument reference.
* `compute_type` - (Required) The type of the compute resource used to execute the "containerAction". Possible values are: ACU_1 (vCPU=4, memory=16GiB) or ACU_2 (vCPU=8, memory=32GiB).
* `volume_size_in_gb` - (Required) The size (in GB) of the persistent storage available to the resource instance used to execute the "containerAction" (min: 1, max: 50).


<a name="variable"><a/> The `variable` argument reference.
* `name` - (Required) The name of the variable.
* `string_value` - (Optional) The value of the variable as a string.
* `double_value` - (Optional) The value of the variable as a double (numeric).
* `dataset_content_version_value` - Object (Optional) Only one. The value of the variable as a structure that specifies a data set content version.
    * `dataset_name` - (Required) The name of the data set whose latest contents are used as input to the notebook or application.
* `output_file_uri_value` - Object (Optional) Only one. The value of the variable as a structure that specifies an output file URI.
    * `file_name` - (Required). The URI of the location where data set contents are stored, usually the URI of a file in an S3 bucket.



<a name="query_action"><a/> The `query_action` argument reference.
* `sql_query` - (Required) A SQL query string.
* [`filter`](#filter) - Object (Optional) Can be many. Pre-filters applied to message data.

<a name="filter"><a/> The `filter` argument reference.
* `delta_time` - Object (Optional) Only one. Used to limit data to that which has arrived since the last execution of the action.
    * `offset_seconds` - (Required) The number of seconds of estimated "in flight" lag time of message data. When you create data set contents using message data from a specified time frame, some message data may still be "in flight" when processing begins, and so will not arrive in time to be processed. Use this field to make allowances for the "in flight" time of your message data, so that data not processed from a previous time frame will be included with the next time frame. Without this, missed message data would be excluded from processing during the next time frame as well, because its timestamp places it within the previous time frame.
    * `time_expression` - (Required). An expression by which the time of the message data may be determined. This may be the name of a timestamp field, or a SQL expression which is used to derive the time the message data was generated.


<a name="content_delivery_rule"><a/> The `content_delivery_rule` argument reference.
* `entry_name` - (Optional) The name of the data set content delivery rules entry.
* [`destination`](#destination) - Object (Optional) Only one. The destination to which data set contents are delivered.

<a name="destination"><a/> The `destination` argument reference.
* [`iotevents_destination`](#iotevents_destination) - Object (Optional) Only one. Configuration information for delivery of data set contents to AWS IoT Events.
* [`s3_destination`](#s3_destination) - Object (Optional) Only one. Configuration information for delivery of data set contents to Amazon S3.

<a name="iotevents_destination"><a/> The `iotevents_destination` argument reference.
* `input_name` - (Required) The name of the AWS IoT Events input to which data set contents are delivered.
* `role_arn` - (Required) The ARN of the role which grants AWS IoT Analytics permission to deliver data set contents to an AWS IoT Events input.

<a name="s3_destination"><a/> The `s3_destination` argument reference.
* `bucket` - (Required) The name of the Amazon S3 bucket to which data set contents are delivered.
* `key` - (Required) The key of the data set contents object. Each object in an Amazon S3 bucket has a key that is its unique identifier within the bucket (each object in a bucket has exactly one key).
* `role_arn` - (Required) The ARN of the role which grants AWS IoT Analytics permission to interact with your Amazon S3 and AWS Glue resources.
* `glue_configuration` - Object (Optional) Only one. Configuration information for coordination with the AWS Glue ETL (extract, transform and load) service.
    * `database_name` - (Required) The name of the database in your AWS Glue Data Catalog in which the tableis located. (An AWS Glue Data Catalog database contains Glue Data tables.)
    * `table_name` - (Required) The name of the table in your AWS Glue Data Catalog which is used to perform the ETL (extract, transform and load) operations. (An AWS Glue Data Catalog table contains partitioned data and descriptions of data sources and targets.)


<a name="trigger"><a/> The `trigger` argument reference.
* `dataset` - Object (Optional) Only one. The data set whose content creation triggers the creation of this data set's contents.
    * `name` - (Required) The name of the data set whose content generation triggers the new data set.
* `schedule` - Object (Optional) Only one. The "Schedule" when the trigger is initiated.
    * `expression` - (Required) The expression that defines when to trigger an update. For more information, see Schedule Expressions for Rules (https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html) in the Amazon CloudWatch Events User Guide.

<a name="retention_period"><a/> The `retention_period` argument reference.
* `number_of_days` - (Optional) The number of days that message data is kept.
* `unlimited` - (Optional) If true, message data is kept indefinitely.

<a name="versioning_configuration"><a/> The `versioning_configuration` argument reference.
* `max_versions` - (Required) How many versions of data set contents will be kept. The "unlimited" parameter must be false.
* `unlmited` - (Required) If true, unlimited versions of data set contents will be kept.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the dataset
* `arn` - The ARN of the dataset.

## Import

IoTAnalytics Dataset can be imported using the `name`, e.g.

```
$ terraform import aws_iotanalytics_dataset.dataset <name>
```
