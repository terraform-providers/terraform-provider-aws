---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_instance"
description: |-
  Provides details about a specific Connect Instance.
---

# Resource: aws_connect_instance

Provides an Amazon Connect instance resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

```hcl
resource "aws_connect_instance" "foo" {
  identity_management_type = "CONNECT_MANAGED"
  instance_alias           = "resource-test-terraform-connect"
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
```

## Argument Reference


The following arguments are supported:

* `identity_management_type` - (Optional) Specifies The identity management type attached to the instance. Defaults to `CONNECT_MANAGED`. Allowed Values are: `SAML`, `CONNECT_MANAGED`, `EXISTING_DIRECTORY`.
* `directory_id` - (Optional) The identifier for the directory if identity_management_type is `EXISTING_DIRECTORY`.
* `instance_alias` - (Optional) Specifies the name of the instance
* `inbound_calls_enabled` - (Optional) Specifies Whether inbound calls are enabled. Defaults to `true`
* `outbound_calls_enabled` -  (Optional) Specifies Whether outbound calls are enabled. * `inbound_calls_enabled` - (Optional) Specifies Whether inbound calls are enabled. Defaults to `true`
* `early_media_enabled` - (Optional) Specifies Whether early media for outbound calls is enabled . Defaults to `true` if outbound calls is enabled
* `contact_flow_logs_enabled` - (Optional) Specifies Whether contact flow logs are enabled. Defaults to `false`
* `contact_lens_enabled` - (Optional) Specifies Whether contact lens is enabled. Defaults to `true`
* `auto_resolve_best_voices` - (Optional) Specifies Whether auto resolve best voices is enabled. Defaults to `true`
* `use_custom_tts_voices` - (Optional) Specifies Whether use custom tts voices is enabled. Defaults to `false`

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when creating the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `created_time` - Specifies When the instance was created.
* `arn` - The Amazon Resource Name (ARN) of the instance.
* `status` - Specifies The state of the instance.
* `service_role` - The service role of the instance.
