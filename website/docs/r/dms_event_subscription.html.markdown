---
subcategory: "Database Migration Service (DMS)"
layout: "aws"
page_title: "AWS: aws_dms_event_subscription"
description: |-
  Provides a DMS (Data Migration Service) event subscription resource.
---

# Resource: aws_dms_event_subscription

Provides a DMS (Data Migration Service) event subscription resource.

## Example Usage

```hcl
resource "aws_dms_event_subscription" "example" {
  enabled          = true
  event_categories = ["creation", "failure"]
  name             = "my-favorite-event-subscription"
  sns_topic_arn    = aws_sns_topic.example.arn
  source_ids       = [aws_dms_replication_task.example.replication_task_id]
  source_type      = "replication-task"

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of event subscription.
* `enabled` - (Optional, Default: true) Whether the event subscription should be enabled.
* `event_categories` - (Optional) List of event categories to listen for, see `DescribeEventCategories` for a canonical list.
* `source_type` - (Optional, Default: all events) Type of source for events. Valid values: `replication-instance` or `replication-task`
* `source_ids` - (Required) Ids of sources to listen to.
* `sns_topic_arn` - (Required) SNS topic arn to send events on.

## Timeouts

`aws_dms_event_subscription` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `10m`) Used for creating event subscriptions.
- `update` - (Default `10m`) Used for event subscription modifications.
- `delete` - (Default `10m`) Used for destroying event descriptions.

## Import

Event subscriptions can be imported using the `name`, e.g.

```
$ terraform import aws_dms_event_subscription.test my-awesome-event-subscription
```
