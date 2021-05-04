---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_rule"
description: |-
  Get information on a Cloudwatch Event rule.
---

# Data Source: aws_cloudwatch_event_rule

Use this data source to get information about an AWS Cloudwatch Event rule

## Example Usage

```hcl
data "aws_cloudwatch_event_rule" "example" {
  name = "MyEventRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Cloudwatch Event rule

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Cloudwatch Event rule
* `description` - The description of the rule.
* `event_pattern` - The event pattern of the rule. For more information, see Events and Event Patterns (https://docs.aws.amazon.com/eventbridge/latest/userguide/eventbridge-and-event-patterns.html)
* `managed_by` - If the rule was created on behalf of your account by an AWS service, this field displays the principal name of the service that created the rule.
* `role_arn` - The Amazon Resource Name (ARN) of the role that is used for target invocation.
* `schedule_expression` - The scheduling expression.
* `state` - The state of the rule.
