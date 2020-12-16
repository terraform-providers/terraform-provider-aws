---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance"
description: |-
  Provides details about a specific Connect Instance.
---

# Data Source: aws_connect_instance

Provides details about a specific Amazon Connect Instance.

## Example Usage
By instance_alias

```hcl
data "aws_connect_instance" "foo" {
  instance_alias = "foo"
}
```

By instance_id

```hcl
data "aws_connect_instance" "foo" {
  instance_id = "97afc98d-101a-ba98-ab97-ae114fc115ec"
}
```

## Argument Reference

~> **NOTE:** One of either `instance_id` or `instance_alias` is required.

The following arguments are supported:

* `instance_id` - (Optional) Returns information on a specific connect instance by id

* `instance_alias` - (Optional) Returns information on a specific connect instance by alias

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_time` - Specifies When the instance was created.
* `arn` - The Amazon Resource Name (ARN) of the instance.
* `identity_management_type` - Specifies The identity management type attached to the instance.
* `inbound_calls_enabled` - Specifies Whether inbound calls are enabled.
* `outbound_calls_enabled` - Specifies Whether outbound calls are enabled.
* `early_media_enabled` - Specifies Whether early media for outbound calls is enabled .
* `contact_flow_logs_enabled` - Specifies Whether contact flow logs are enabled.
* `contact_lens_enabled` - Specifies Whether contact lens is enabled.
* `auto_resolve_best_voices` - Specifies Whether auto resolve best voices is enabled.
* `use_custom_tts_voices` - Specifies Whether use custom tts voices is enabled.
* `status` - Specifies The state of the instance.
* `service_role` - The service role of the instance.
