---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_contact_flow"
description: |-
  Provides details about a specific Amazon Connect Contact Flow.
---

# Resource: aws_connect_contact_flow

Provides an Amazon Connect contact flow resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

This resource embeds or references Contact Flows specified in Amazon connect Contact Flow Language. For more information see
[Amazon Connect Flow language](https://docs.aws.amazon.com/connect/latest/adminguide/flow-language.html)

~> **NOTE:** Due to The behaviour of Amazon Connect you cannot delete contact flows [Create a new contact flow](https://docs.aws.amazon.com/connect/latest/adminguide/create-contact-flow.html), instead the recommendation is to prefix the Contact Flow with `zzTrash`. This resource will automatically apply this logic by renaming the Contact Flow to `zzTrash`+original name+`timestamp` if the resource is removed or disappears.

## Example Usage

### Basic

```hcl
resource "aws_connect_contact_flow" "foo" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "Test"
  description = "Test Contact Flow Description"
  type        = "CONTACT_FLOW"
  content     = <<JSON
	{
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Type": "MessageParticipant",
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Parameters": {
					"Text": "Thanks for calling the sample flow!"
				}
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Transitions": {},
				"Parameters": {}
			}
		]
	}
	JSON
  tags = map(
    "Name", "Test Contact Flow",
    "Application", "Terraform",
    "Method", "Create"
  )
}
```

### With External Content

```hcl
resource "aws_connect_contact_flow" "foo" {
  instance_id  = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name         = "Test"
  description  = "Test Contact Flow Description"
  type         = "CONTACT_FLOW"
  filename     = "connect_flow.json"
  content_hash = filebase64sha256("connect_flow.json")
  tags = map(
    "Name", "Test Contact Flow",
    "Application", "Terraform",
    "Method", "Create"
  )
}
```

## Argument Reference


The following arguments are supported:

* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `name` - (Required) Specifies the name of the Contact Flow.
* `description` - (Optional) Specifies the description of the Contact Flow.
* `type` - (Optional) Specifies the type of the Contact Flow. Defaults to `CONTACT_FLOW`. Allowed Values are: `CONTACT_FLOW`, `CONTACT_FLOW`, `CUSTOMER_QUEUE`, `CUSTOMER_HOLD`, `CUSTOMER_WHISPER`, `AGENT_HOLD`, `AGENT_WHISPER`, `OUTBOUND_WHISPER`, `AGENT_TRANSFER`, `QUEUE_TRANSFER`.
* `content` - (Optional) Specifies the content of the Contact Flow, provided as a JOSN string, written in Amazon connect Contact Flow Language. If defined, the `filename` argument cannot be used.
* `filename` - (Optional) The path to the Contact Flow source within the local filesystem. Conflicts with `content`.
* `content_hash` - (Optional) Used to trigger updates. Must be set to a base64-encoded SHA256 hash of the Contact Flow source specified with `filename`. The usual way to set this is filebase64sha256("mycontact_flow.json") (Terraform 0.11.12 and later) or base64sha256(file("mycontact_flow.json")) (Terraform 0.11.11 and earlier), where "mycontact_flow.json" is the local filename of the Contact Flow source.
* `tags` - (Optional) A map of tags to assign to the Contact Flow.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 1 min) Used when creating the Contact Flow.
* `update` - (Defaults to 1 min) Used when updating the Contact Flow.
* `delete` - (Defaults to 1 min) Used when deleting the Contact Flow.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `contact_flow_id` - Specifies the unique reference to the Contact Flow.
* `arn` - The Amazon Resource Name (ARN) of the Contact Flow.
