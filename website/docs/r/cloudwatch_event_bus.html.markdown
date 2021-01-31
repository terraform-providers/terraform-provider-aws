---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus"
description: |-
  Provides an EventBridge event bus resource.
---

# Resource: aws_cloudwatch_event_bus

Provides an EventBridge event bus resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```hcl
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new event bus. The names of custom event buses can't contain the / character. Please note that a partner event bus is not supported at the moment.
* `tags` - (Optional)  A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the event bus.


## Import

EventBridge event buses can be imported using the `name`, e.g.

```console
$ terraform import aws_cloudwatch_event_bus.messenger chat-messages
```
