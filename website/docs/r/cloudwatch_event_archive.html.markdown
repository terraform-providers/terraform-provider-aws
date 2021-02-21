---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_archive"
description: |-
  Provides an EventBridge event archive resource.
---

# Resource: aws_cloudwatch_event_archive

Provides an EventBridge event archive resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```hcl
resource "aws_cloudwatch_event_bus" "order" {
  name = "orders"
}

resource "aws_cloudwatch_event_archive" "order" {
  name             = "order-archive"
  event_source_arn = aws_cloudwatch_event_bus.order.arn
}
```

## Example all optional arguments

```hcl
resource "aws_cloudwatch_event_bus" "order" {
  name = "orders"
}

resource "aws_cloudwatch_event_archive" "order" {
  name             = "order-archive"
  description      = "Archived events from order service"
  event_source_arn = aws_cloudwatch_event_bus.order.arn
  retention_days   = 7
  event_pattern    = <<PATTERN
{
  "source": ["company.team.order"]
}
PATTERN
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new event archive. The archive name cannot exceed 48 characters.
* `event_source_arn` - (Required) Event bus source ARN from where these events should be archived.
* `description` - (Optional) The description of the new event archive.
* `event_pattern` - (Optional) Instructs the new event archive to only capture events matched by this pattern. By default, it attempts to archive every event received in the `event_source_arn`.
* `retention_days` - (Optional) The maximum number of days to retain events in the new event archive. By default, it archives indefinitely.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the event archive.

## Import

Event Archive can be imported using their name, for example

```bash
terraform import aws_cloudwatch_event_archive.imported_event_archive order-archive
```
