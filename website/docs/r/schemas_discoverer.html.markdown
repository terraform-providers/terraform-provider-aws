---
subcategory: "EventBridge Schemas"
layout: "aws"
page_title: "AWS: aws_schemas_discoverer"
description: |-
  Provides an EventBridge Schema Discoverer resource.
---

# Resource: aws_schemas_discoverer

Provides an EventBridge Schema Discoverer resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```terraform
resource "aws_cloudwatch_event_bus" "messenger" {
  name = "chat-messages"
}

resource "aws_schemas_discoverer" "test" {
  source_arn  = aws_cloudwatch_event_bus.messenger.arn
  description = "Auto discover event schemas"
}
```

## Argument Reference

The following arguments are supported:

* `source_arn` - (Required) The ARN of the event bus to discover event schemas on.
* `description` - (Optional) The description of the discoverer. Maximum of 256 characters.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the discoverer.
* `id` - The ID of the discoverer.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

EventBridge discoverers can be imported using the `id`, e.g.

```console
$ terraform import aws_schemas_discoverer.test 123
```
