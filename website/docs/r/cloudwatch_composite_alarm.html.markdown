---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_composite_alarm"
description: |-
  Provides a CloudWatch Composite Alarm resource.
---

# Resource: aws_cloudwatch_composite_alarm

Provides a CloudWatch Composite Alarm resource.

~> **NOTE:** An alarm (composite or metric) cannot be destroyed when there are other composite alarms depending on it. This can lead to a cyclical dependency on update, as Terraform will unsuccessfully attempt to destroy alarms before updating the rule. Consider using `depends_on`, references to alarm names, and two-stage updates.

## Example Usage

```hcl
resource "aws_cloudwatch_composite_alarm" "example" {
  alarm_description = "This is a composite alarm!"
  alarm_name        = "example-composite-alarm"

  alarm_actions = aws_sns_topic.example.arn
  ok_actions    = aws_sns_topic.example.arn

  alarm_rule = <<EOF
ALARM(${aws_cloudwatch_metric_alarm.alpha.alarm_name}) OR
ALARM(${aws_cloudwatch_metric_alarm.bravo.alarm_name})
EOF
}
```

## Argument Reference

* `actions_enabled` - (Optional, Forces new resource) Indicates whether actions should be executed during any changes to the alarm state of the composite alarm. Defaults to `true`.
* `alarm_actions` - (Optional) The set of actions to execute when this alarm transitions to the `ALARM` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `alarm_description` - (Optional) The description for the composite alarm.
* `alarm_name` - (Required) The name for the composite alarm. This name must be unique within the region.
* `alarm_rule` - (Required) An expression that specifies which other alarms are to be evaluated to determine this composite alarm's state. For syntax, see [Creating a Composite Alarm](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Create_Composite_Alarm.html). The maximum length is 10240 characters.
* `insufficient_data_actions` - (Optional) The set of actions to execute when this alarm transitions to the `INSUFFICIENT_DATA` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `ok_actions` - (Optional) The set of actions to execute when this alarm transitions to an `OK` state from any other state. Each action is specified as an ARN. Up to 5 actions are allowed.
* `tags` - (Optional) A map of tags to associate with the alarm. Up to 50 tags are allowed.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the composite alarm.
* `id` - The ID of the composite alarm resource, which is equivalent to its `alarm_name`.

## Import

Use the `alarm_name` to import a CloudWatch Composite Alarm. For example:

```
$ terraform import aws_cloudwatch_composite_alarm.test my-alarm
```
