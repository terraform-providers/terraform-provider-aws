---
layout: "aws"
page_title: "AWS: ses_event_destination"
sidebar_current: "docs-aws-resource-ses-event-destination"
description: |-
  Provides an SES event destination
---

# aws_ses_event_destination

Provides an SES event destination

## Example Usage

### CloudWatch Destination

```hcl
resource "aws_ses_event_destination" "cloudwatch" {
  name                   = "event-destination-cloudwatch"
  configuration_set_name = "${aws_ses_configuration_set.example.name}"
  enabled                = true
  matching_types         = ["bounce", "send"]

  cloudwatch_destination = {
    default_value  = "default"
    dimension_name = "dimension"
    value_source   = "emailHeader"
  }
}
```

### Kinesis Destination

```hcl
resource "aws_ses_event_destination" "kinesis" {
  name                   = "event-destination-kinesis"
  configuration_set_name = "${aws_ses_configuration_set.example.name}"
  enabled                = true
  matching_types         = ["bounce", "send"]

  kinesis_destination = {
    stream_arn = "${aws_kinesis_firehose_delivery_stream.example.arn}"
    role_arn   = "${aws_iam_role.example.arn}"
  }
}
```

### SNS Destination

```hcl
resource "aws_ses_event_destination" "sns" {
  name                   = "event-destination-sns"
  configuration_set_name = "${aws_ses_configuration_set.example.name}"
  enabled                = true
  matching_types         = ["bounce", "send"]

  sns_destination {
    topic_arn = "${aws_sns_topic.example.arn}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the event destination
* `configuration_set_name` - (Required) The name of the configuration set
* `enabled` - (Optional) If true, the event destination will be enabled
* `matching_types` - (Required) A list of matching types. May be any of `"send"`, `"reject"`, `"bounce"`, `"complaint"`, `"delivery"`, `"open"`, `"click"`, or `"renderingFailure"`.
* `cloudwatch_destination` - (Optional) CloudWatch destination for the events
* `kinesis_destination` - (Optional) Send the events to a kinesis firehose destination
* `sns_destination` - (Optional) Send the events to an SNS Topic destination

~> **NOTE:** You can specify `"cloudwatch_destination"` or `"kinesis_destination"` but not both

### cloudwatch_destination Argument Reference

* `default_value` - (Required) The default value for the event
* `dimension_name` - (Required) The name for the dimension
* `value_source` - (Required) The source for the value. It can be either `"messageTag"` or `"emailHeader"`

### kinesis_destination Argument Reference

* `stream_arn` - (Required) The ARN of the Kinesis Stream
* `role_arn` - (Required) The ARN of the role that has permissions to access the Kinesis Stream

### sns_destination Argument Reference

* `topic_arn` - (Required) The ARN of the SNS topic
