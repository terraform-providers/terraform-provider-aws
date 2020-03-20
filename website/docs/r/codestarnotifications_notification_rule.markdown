---
subcategory: "CodeStar Notifications"
layout: "aws"
page_title: "AWS: aws_codestarnotifications_notification_rule"
description: |-
  Provides a CodeStar Notifications Rule
---

# Resource: aws_codestarnotifications_notification_rule

Provides a CodeStar Notifications Rule.

## Example Usage

```hcl
resource "aws_codecommit_repository" "code" {
  repository_name = "example-code-repo"
}

resource "aws_sns_topic" "notif" {
  name = "notification"
}

data "aws_iam_policy_document" "notif_access" {
  statement {
    actions = ["sns:Publish"]

    principals {
      type        = "Service"
      identifiers = ["codestar-notifications.amazonaws.com"]
    }

    resources = [aws_sns_topic.notif.arn]
  }
}

resource "aws_sns_topic_policy" "default" {
  arn    = aws_sns_topic.notif.arn
  policy = data.aws_iam_policy_document.notif_access.json
}

resource "aws_codestarnotifications_notification_rule" "commits" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]

  name     = "example-code-repo-commits"
  resource = aws_codecommit_repository.code.arn

  target {
    address = aws_sns_topic.notif.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `detail_type` - (Required) The level of detail to include in the notifications for this resource. Possible values are `BASIC` and `FULL`.
* `event_type_ids` - (Required) A list of event types associated with this notification rule.
  For list of allowed events see [here](https://docs.aws.amazon.com/codestar-notifications/latest/userguide/concepts.html#concepts-api).
* `name` - (Required) The name of notification rule.
* `resource` - (Required) The ARN of the resource to associate with the notification rule.
* `status` - (Optional) The status of the notification rule. Possible values are `ENABLED` and `DISABLED`, default is `ENABLED`.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `target` - (Optional) Configuration blocks containing notification target information. Can be specified multiple times. At least one target must be specified on creation.

An `target` block supports the following arguments:

* `address` - (Required) The ARN of notification rule target. For example, a SNS Topic ARN.
* `type` - (Optional) The type of the notification target. Default value is `SNS`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The codestar notification rule ARN.
* `arn` - The codestar notification rule ARN.

## Import

CodeStar notification rule can be imported using the ARN, e.g.

```
$ terraform import aws_codestarnotification_rule.foo arn:aws:codestar-notifications:us-west-1:0123456789:notificationrule/2cdc68a3-8f7c-4893-b6a5-45b362bd4f2b
```
