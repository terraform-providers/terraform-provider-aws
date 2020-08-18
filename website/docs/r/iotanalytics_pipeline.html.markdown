---
subcategory: "IoTAnalytics"
layout: "aws"
page_title: "AWS: aws_iotanalytics_pipeline"
sidebar_current: "docs-aws-resource-iotanalytics-pipeline"
description: |-
    Creates and manages an AWS IoTAnalytics Pipeline
---

# Resource: aws_iotanalytics_pipeline

## Example Usage

```hcl
resource "aws_iotanalytics_pipeline" "pipeline" {
  name = "test_pipeline_%[1]s"
 
  pipeline_activity {
	  channel {
		name = "channel_activity"
		channel_name = "channel_name"
		next_activity = "datastore_activity"
	  }
  }

  pipeline_activity {
	datastore {
		name = "datastore_activity"
		datastore_name = "datastore_name"
	}
  }
}
```

## Argument Reference

* `name` - (Required) The name of the input.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the pipeline.
* `pipeline_activity` - (Required) Object, Can be multiple limited to 25. A list of "PipelineActivity" objects. Activities perform transformations on your messages, such as removing, renaming or adding message attributes; filtering messages based on attribute values; invoking your Lambda functions on messages for advanced processing; or performing mathematical transformations to normalize device data. The list can be 2-25 PipelineActivity objects and must contain both a channel and a datastore activity. Each entry in the list must contain only one activity. It must be said that ordering of pipeline activities matter. Channel activity should always be first and datastore last. Other activities should be ordered depending on `next_activity` field of previous activity in such way that name of activity should be equals to value in `next_activity` field of previous activity.

`pipeline_activity` takes such arguments:

* [`add_attributes`](#add_attributes) - (Optional) Only one. Adds other attributes based on existing attributes in the message.
* [`remove_attributes`](#remove_attributes) - (Optional) Only one. Removes attributes from a message.
* [`select_attributes`](#select_attributes) - (Optional) Only one. Creates a new message using only the specified attributes from the original message.
* [`channel`](#channel) - (Optional) Only one. Determines the source of the messages to be processed.
* [`datastore`](#datastore) - (Optional) Only one. Specifies where to store the processed message data.
* [`device_registry_enrich`](#device_registry_enrich) - (Optional) Only one. Adds data from the AWS IoT device registry to your message.
* [`device_shadow_enrich`](#device_shadow_enrich) - (Optional) Only one. Adds information from the AWS IoT Device Shadows service to a message.
* [`filter`](#filter) - (Optional) Only one. Filters a message based on its attributes.
* [`lambda`](#lambda) - (Optional) Only one. Runs a Lambda function to modify the message.
* [`math`](#math) - (Optional) Only one. Computes an arithmetic expression using the message's attributes and adds it to the message.

<a name="add_attributes"><a/> The `add_attributes` argument reference.
* `name` - (Required). The name of the `add_attributes` activity.
* `attributes` - (Requied) Map. A list of 1-50 "AttributeNameMapping" objects that map an existing attribute to a new attribute. The existing attributes remain in the message, so if you want to remove the originals, use `remove_attributes`.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="remove_attributes"><a/> The `remove_attributes` argument reference.
* `name` - (Required). The name of the `add_attributes` activity.
* `attributes` - (Required). A list of 1-50 attributes to remove from the message.
* `next_activity` - (Optional) List. The next activity in the pipeline.

<a name="select_attributes"><a/> The `select_attributes` argument reference.
* `name` - (Required). The name of the `select_attributes` activity.
* `attributes` - (Required) List. A list of the attributes to select from the message.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="channel"><a/> The `channel` argument reference.
* `name` - (Required). The name of the `channel` activity.
* `channel_name` - (Required). The name of the channel from which the messages are processed.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="datastore"><a/> The `datastore` argument reference.
* `name` - (Required). The name of the `datastore` activity.
* `datastore_name` - (Required). The name of the data store where processed messages are stored.

<a name="device_registry_enrich"><a/> The `device_registry_enrich` argument reference.
* `name` - (Required). The name of the `device_registry_enrich` activity.
* `attribute` - (Required). The name of the attribute that is added to the message.
* `role_arn` - (Required). The ARN of the role that allows access to the device's registry information.
* `thing_name` - (Required). The name of the IoT device whose registry information is added to the message.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="device_shadow_enrich"><a/> The `device_shadow_enrich` argument reference.
* `name` - (Required). The name of the `device_shadow_enrich` activity.
* `attribute` - (Required). The name of the attribute that is added to the message.
* `role_arn` - (Required). The ARN of the role that allows access to the device's registry information.
* `thing_name` - (Required). The name of the IoT device whose registry information is added to the message.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="filter"><a/> The `filter` argument reference.
* `name` - (Required). The name of the `filter` activity.
* `filter` - (Required). An expression that looks like a SQL WHERE clause that must return a Boolean value.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="lambda"><a/> The `lambda` argument reference.
* `name` - (Required). The name of the `lambda` activity.
* `lambda_name` - (Required). The name of the Lambda function that is run on the message.
* `batch_size` - (Required). The number of messages passed to the Lambda function for processing. The AWS Lambda function must be able to process all of these messages within five minutes, which is the maximum timeout duration for Lambda functions.
* `next_activity` - (Optional). The next activity in the pipeline.

<a name="math"><a/> The `math` argument reference.
* `name` - (Required). The name of the `math` activity.
* `math` - (Required). An expression that uses one or more existing attributes and must return an integer value.
* `attribute` - (Required). The name of the attribute that contains the result of the math operation.
* `next_activity` - (Optional). The next activity in the pipeline.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the pipeline
* `arn` - The ARN of the pipeline.

## Import

IoTAnalytics Channel can be imported using the `name`, e.g.

```
$ terraform import aws_iotanalytics_pipeline.pipeline <name>
```
