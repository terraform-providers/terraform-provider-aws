---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_core_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Core Definition
---

# Resource: aws_greengrass_core_definition

## Example Usage

```hcl
resource "aws_greengrass_core_definition" "test" {
	name = "core_definition_%[1]s"
	core_definition_version {
		core {
			certificate_arn = "aws_iot_certificate arn"
			id = "core_id"
			sync_shadow = false
			thing_arn = "aws_iot_thing arn"
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the core definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the core definition.
* `core_definition_version` - (Optional) Object. Information about a core definition version.

The `core_definition_version` object has such arguments.
* `core` - (Optional) List of Object. A list of references to cores in this version, with their corresponding configuration settings.

The `core` object has such arguments:
* `certificate_arn` - (Required) The ARN of the certificate associated with the core.
* `id` - (Required) A descriptive or arbitrary ID for the core. This value must be unique within the core definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `sync_shadow` - (Optional) Default `false`. If true, the core's local shadow will be automatically synced with the cloud.
* `thing_arn` - (Required) The thing ARN of the core.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `core_definition_version.arn` - The ARN of latest core definition version

## Environment variables
If you use `core_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Core Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_core_definition.definition <core_definition_id>
``` 