---
layout: "aws"
page_title: "AWS: aws_pinpoint_email_channel"
sidebar_current: "docs-aws-resource-pinpoint-email-channel"
description: |-
  Provides a Pinpoint SMS Channel resource.
---

# aws_pinpoint_email_channel

Provides a Pinpoint SMS Channel resource.

## Example Usage

```hcl
resource "aws_pinpoint_email_channel" "email" {
  application_id = "${aws_pinpoint_app.app.application_id}"
  from_address   = "user@example.com"
  identity       = "${aws_ses_domain_identity.identity.arn}"
  role_arn       = "${aws_iam_role.role.arn}"
}

resource "aws_pinpoint_app" "app" {}

resource "aws_ses_domain_identity" "identity" {
  domain = "example.com"
}

resource "aws_iam_role" "role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "role_policy" {
  name = "role_policy"
  role = "${aws_iam_role.role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "mobileanalytics:PutEvents",
      "mobileanalytics:PutItems"
    ],
    "Effect": "Allow",
    "Resource": [
      "*"
    ]
  }
}
EOF
}
```


## Argument Reference

The following arguments are supported:

* `application_id` - (Required) The application ID.
* `enabled` - (Optional) Whether the channel is enabled or disabled. Defaults to `true`.
* `from_address` - (Required) The email address used to send emails from.
* `identity` - (Required) The ARN of an identity verified with SES.
* `role_arn` - (Required) The ARN of an IAM Role used to submit events to Mobile Analytics' event ingestion service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `messages_per_second` - Messages per second that can be sent.

## Import

Pinpoint Email Channel can be imported using the `application-id`, e.g.

```
$ terraform import aws_pinpoint_email_channel.email application-id
```
