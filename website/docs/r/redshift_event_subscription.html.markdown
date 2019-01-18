---
layout: "aws"
page_title: "AWS: aws_redshift_event_subscription"
sidebar_current: "docs-aws-resource-redshift-event-subscription"
description: |-
  Provides a Redshift event subscription resource.
---

# aws_redshift_event_subscription

Provides a Redshift event subscription resource.

## Example Usage

```hcl
resource "aws_redshift_cluster" "default" {
  cluster_identifier = "default"
  database_name      = "default"

  # ...
}

resource "aws_sns_topic" "default" {
  name = "redshift-events"
}

resource "aws_redshift_event_subscription" "default" {
  name      = "redshift-event-sub"
  sns_topic = "${aws_sns_topic.default.arn}"

  source_type = "cluster"
  source_ids  = ["${aws_redshift_cluster.default.id}"]

  severity = "INFO"

  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]

  tags = {
    Name = "default"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Redshift event subscription.
* `sns_topic_arn` - (Required) The ARN of the SNS topic to send events to.
* `source_ids` - (Optional) A list of identifiers of the event sources for which events will be returned. If not specified, then all sources are included in the response. If specified, a source_type must also be specified.
* `source_type` - (Optional) The type of source that will be generating the events. Valid options are `cluster`, `cluster-parameter-group`, `cluster-security-group`, or `cluster-snapshot`. If not set, all sources will be subscribed to.
* `severity` - (Optional) The event severity to be published by the notification subscription. Valid options are `INFO` or `ERROR`.
* `event_categories` - (Optional) A list of event categories for a SourceType that you want to subscribe to. See https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-event-notifications.html or run `aws redshift describe-event-categories`.
* `enabled` - (Optional) A boolean flag to enable/disable the subscription. Defaults to true.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes

The following additional atttributes are provided:

* `id` - The name of the Redshift event notification subscription
* `customer_aws_id` - The AWS customer account associated with the Redshift event notification subscription

## Import

Redshift Event Subscriptions can be imported using the `name`, e.g.

```
$ terraform import aws_redshift_event_subscription.default redshift-event-sub
```
