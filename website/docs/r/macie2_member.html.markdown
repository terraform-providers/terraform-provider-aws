---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_member"
description: |-
  Provides a resource to manage an Amazon Macie Member.
---

# Resource: aws_macie2_member

Provides a resource to manage an [Amazon Macie Member](https://docs.aws.amazon.com/macie/latest/APIReference/members-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_member" "example" {
  account_id                            = "AWS ACCOUNT ID"
  email                                 = "EMAIL"
  invite                                = true
  invitation_message                    = "Message of the invitation"
  invitation_disable_email_notification = true
  depends_on                            = [aws_macie2_account.example]
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The AWS account ID for the account.
* `email` - (Required) The email address for the account.
* `tags` - (Optional) A map of key-value pairs that specifies the tags to associate with the account in Amazon Macie.
* `status` - (Optional) Specifies the status for the account. To enable Amazon Macie and start all Macie activities for the account, set this value to `ENABLED`. Valid values are `ENABLED` or `PAUSED`.
* `invite` - (Optional) Send an invitation to a member
* `invitation_message` - (Optional) A custom message to include in the invitation. Amazon Macie adds this message to the standard content that it sends for an invitation.
* `invitation_disable_email_notification` - (Optional) Specifies whether to send an email notification to the root user of each account that the invitation will be sent to. This notification is in addition to an alert that the root user receives in AWS Personal Health Dashboard. To send an email notification to the root user of each account, set this value to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie Member.
* `arn` - The Amazon Resource Name (ARN) of the account.
* `relationship_status` - The current status of the relationship between the account and the administrator account.
* `administrator_account_id` - The AWS account ID for the administrator account.
* `invited_at` - The date and time, in UTC and extended RFC 3339 format, when an Amazon Macie membership invitation was last sent to the account. This value is null if a Macie invitation hasn't been sent to the account.
* `updated_at` - The date and time, in UTC and extended RFC 3339 format, of the most recent change to the status of the relationship between the account and the administrator account.

## Import

`aws_macie2_member` can be imported using the account ID of the member account, e.g.

```
$ terraform import aws_macie2_member.example 123456789012
```
