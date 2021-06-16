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

```terraform
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}
```

```terraform
data "aws_cloudwatch_event_source" "examplepartner" {
  name_prefix = "aws.partner/examplepartner.com"
}

resource "aws_cloudwatch_event_bus" "examplepartner" {
  name              = data.aws_cloudwatch_event_source.examplepartner.name
  event_source_name = data.aws_cloudwatch_event_source.examplepartner.name
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new event bus. The names of custom event buses can't contain the / character. To create a partner event bus, ensure the `name` matches the `event_source_name`.
* `event_source_name` (Optional) The partner event source that the new event bus will be matched with. Must match `name`.
* `tags` - (Optional)  A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the event bus.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

EventBridge event buses can be imported using the `name` (which can also be a partner event source name), e.g.

```console
$ terraform import aws_cloudwatch_event_bus.messenger chat-messages
```
