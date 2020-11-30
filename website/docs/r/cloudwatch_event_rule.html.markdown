---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_rule"
description: |-
  Provides an EventBridge Rule resource.
---

# Resource: aws_cloudwatch_event_rule

Provides an EventBridge Rule resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```hcl
resource "aws_cloudwatch_event_rule" "console" {
  name        = "capture-aws-sign-in"
  description = "Capture each AWS Console Sign In"

  event_pattern = <<EOF
{
  "detail-type": [
    "AWS Console Sign In via CloudTrail"
  ]
}
EOF
}

resource "aws_cloudwatch_event_target" "sns" {
  rule      = aws_cloudwatch_event_rule.console.name
  target_id = "SendToSNS"
  arn       = aws_sns_topic.aws_logins.arn
}

resource "aws_sns_topic" "aws_logins" {
  name = "aws-console-logins"
}

resource "aws_sns_topic_policy" "default" {
  arn    = aws_sns_topic.aws_logins.arn
  policy = data.aws_iam_policy_document.sns_topic_policy.json
}

data "aws_iam_policy_document" "sns_topic_policy" {
  statement {
    effect  = "Allow"
    actions = ["SNS:Publish"]

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }

    resources = [aws_sns_topic.aws_logins.arn]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the rule. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `schedule_expression` - (Optional) The scheduling expression. For example, `cron(0 20 * * ? *)` or `rate(5 minutes)`. At least one of `schedule_expression` or `event_pattern` is required. Can only be used on the default event bus.
* `event_bus_name` - (Optional) The event bus to associate with this rule. If you omit this, the `default` event bus is used.
* `event_pattern` - (Optional) The event pattern described a JSON object. At least one of `schedule_expression` or `event_pattern` is required. See full documentation of [Events and Event Patterns in EventBridge](https://docs.aws.amazon.com/eventbridge/latest/userguide/eventbridge-and-event-patterns.html) for details.
* `description` - (Optional) The description of the rule.
* `role_arn` - (Optional) The Amazon Resource Name (ARN) associated with the role that is used for target invocation.
* `is_enabled` - (Optional) Whether the rule should be enabled (defaults to `true`).
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the rule.
* `arn` - The Amazon Resource Name (ARN) of the rule.


## Import

EventBridge Rules can be imported using the `event_bus_name/rule_name` (if you omit `event_bus_name`, the `default` event bus will be used), e.g.

```
$ terraform import aws_cloudwatch_event_rule.console example-event-bus/capture-console-sign-in
```
