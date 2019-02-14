---
layout: "aws"
page_title: "AWS: aws_pinpoint_adm_channel"
sidebar_current: "docs-aws-resource-pinpoint-adm-channel"
description: |-
  Provides a Pinpoint ADM Channel resource.
---

# aws_pinpoint_adm_channel

Provides a Pinpoint ADM (Amazon Device Messaging) Channel resource.

~> **Note:** All arguments including the Client ID and Client Secret will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).


## Example Usage

```hcl
resource "aws_pinpoint_app" "app" {}

resource "aws_pinpoint_adm_channel" "channel" {
  application_id = "${aws_pinpoint_app.app.application_id}"
  client_id      = ""
  client_secret  = ""
  enabled        = true
}
```


## Argument Reference

The following arguments are supported:

* `application_id` - (Required) The application ID.
* `client_id` - (Required) Client ID (part of OAuth Credentials) obtained via Amazon Developer Account.
* `client_secret` - (Required) Client Secret (part of OAuth Credentials) obtained via Amazon Developer Account.
* `enabled` - (Optional) Specifies whether to enable the channel. Defaults to `true`.

## Import

Pinpoint ADM Channel can be imported using the `application-id`, e.g.

```
$ terraform import aws_pinpoint_adm_channel.channel application-id
```
