---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_identity_notification_topic"
description: |-
  Setting AWS SES Identity Notification Topic
---

# Resource: aws_ses_identity_notification_topic

Resource for managing SES Identity Notification Topics

## Example Usage

```hcl
resource "aws_ses_identity_notification_topic" "test" {
  topic_arn                = aws_sns_topic.example.arn
  notification_type        = "Bounce"
  identity                 = aws_ses_domain_identity.example.domain
  include_original_headers = true
}
```

## Argument Reference

The following arguments are supported:

* `topic_arn` - (Optional) The Amazon Resource Name (ARN) of the Amazon SNS topic. Can be set to "" (an empty string) to disable publishing.
* `notification_type` - (Required) The type of notifications that will be published to the specified Amazon SNS topic. Valid Values: *Bounce*, *Complaint* or *Delivery*.
* `identity` - (Required) The identity for which the Amazon SNS topic will be set. You can specify an identity by using its name or by using its Amazon Resource Name (ARN).
* `include_original_headers` - (Optional) Whether SES should include original email headers in SNS notifications of this type. *false* by default.

## Import

Identity Notification Topics can be imported using ID of the record. The ID is made up as IDENTITY|TYPE where IDENTITY is the SES Identity and TYPE is the Notification Type.

e.g.

```
example.com|Bounce
```

In this example, `example.com` is the SES Identity and `Bounce` is the Notification Type.

To import the ID above, it would look as follows:

```
$ terraform import aws_ses_identity_notification_topic.test 'example.com|Bounce'
```
