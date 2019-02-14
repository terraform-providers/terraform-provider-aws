---
layout: "aws"
page_title: "AWS: aws_pinpoint_sms_channel"
sidebar_current: "docs-aws-resource-pinpoint-sms-channel"
description: |-
  Provides a Pinpoint SMS Channel resource.
---

# aws_pinpoint_sms_channel

Provides a Pinpoint SMS Channel resource.

## Example Usage

```hcl
resource "aws_pinpoint_sms_channel" "sms" {
  application_id = "${aws_pinpoint_app.app.application_id}"
}

resource "aws_pinpoint_app" "app" {}
```


## Argument Reference

The following arguments are supported:

* `application_id` - (Required) The application ID.
* `enabled` - (Optional) Whether the channel is enabled or disabled. Defaults to `true`.
* `sender_id` - (Optional) Sender identifier of your messages.
* `short_code` - (Optional) The Short Code registered with the phone provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `promotional_messages_per_second` - Promotional messages per second that can be sent.
* `transactional_messages_per_second` - Transactional messages per second that can be sent.

## Import

Pinpoint SMS Channel can be imported using the `application-id`, e.g.

```
$ terraform import aws_pinpoint_sms_channel.sms application-id
```
