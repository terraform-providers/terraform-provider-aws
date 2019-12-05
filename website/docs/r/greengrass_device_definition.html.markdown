---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_device_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Device Definition
---

# Resource: aws_greengrass_device_definition

## Example Usage

```hcl
resource "aws_greengrass_device_definition" "test" {
	name = "device_definition_%[1]s"
	device_definition_version {
		device {
			certificate_arn = "aws_iot_certificate arn"
			id = "device_id"
			sync_shadow = false
			thing_arn = "aws_iot_thing arn"
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the device definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the device definition.
* `device_definition_version` - (Optional) Object. Information about a device definition version.

The `device_definition_version` object has such arguments.
* `device` - (Optional) List of Object. A list of references to devices in this version, with their corresponding configuration settings.

The `device` object has such arguments:
* `certificate_arn` - (Required) The ARN of the certificate associated with the device.
* `id` - (Required) A descriptive or arbitrary ID for the device. This value must be unique within the device definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `sync_shadow` - (Optional) Default `false`. If true, the device's local shadow will be automatically synced with the cloud.
* `thing_arn` - (Required) The thing ARN of the device.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `device_definition_version.arn` - The ARN of latest device definition version

## Environment variables
If you use `device_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Device Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_device_definition.definition <device_definition_id>
``` 