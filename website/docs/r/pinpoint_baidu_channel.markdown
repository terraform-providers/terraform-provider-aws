---
layout: "aws"
page_title: "AWS: aws_pinpoint_baidu_channel"
sidebar_current: "docs-aws-resource-pinpoint-baidu-channel"
description: |-
  Provides a Pinpoint Baidu Channel resource.
---

# aws_pinpoint_baidu_channel

Provides a Pinpoint Baidu Channel resource.

~> **Note:** All arguments including the Api Key and Secret Key will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).


## Example Usage

```hcl
resource "aws_pinpoint_app" "app" {}

resource "aws_pinpoint_baidu_channel" "channel" {
  application_id = "${aws_pinpoint_app.app.application_id}"
  api_key        = ""
  secret_key     = ""
}
```


## Argument Reference

The following arguments are supported:

* `application_id` - (Required) The application ID.
* `enabled` - (Optional) Specifies whether to enable the channel. Defaults to `true`.
* `api_key` - (Required) Platform credential API key from Baidu.
* `secret_key` - (Required) Platform credential Secret key from Baidu.

## Import

Pinpoint Baidu Channel can be imported using the `application-id`, e.g.

```
$ terraform import aws_pinpoint_baidu_channel.channel application-id
```
