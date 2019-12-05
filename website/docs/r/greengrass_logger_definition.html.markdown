---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_logger_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Logger Definition
---

# Resource: aws_greengrass_logger_definition

## Example Usage

```hcl
resource "aws_greengrass_logger_definition" "test" {
	name = "logger_definition_%[1]s"
	logger_definition_version {
		logger {
			component = "GreengrassSystem"
			id = "test_id"
			type = "FileSystem"
			level = "DEBUG"
			space = 3	
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the logger definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the logger definition.
* `logger_definition_version` - (Optional) Object. Information about a logger definition version.

The `logger_definition_version` object has such arguments.
* `logger` - (Optional) List of Object. A list of loggers.

The `logger` object has such arguments:
* `component` - (Required) String. The component that will be subject to logging. This argument can accept such values: Crash, GreengrassSystem, Lambda.
* `id` - (Required) String. A descriptive or arbitrary ID for the core. This value must be unique within the core definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `type` - (Required) String. The type of log output which will be used. This argument can accept such values: AWSCloudWatch, FileSystem.
* `level` - (Required) String. The level of the logs. This argument can accept such values: TRACE, DEBUG, INFO, WARN, ERROR, FATAL.
* `space` - (Optional) Number. The amount of file space, in KB, to use if the local file system is used for logging purposes.


## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `latest_definition_version_arn` - The ARN of latest logger definition version

## Environment variables
If you use `logger_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Logger Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_logger_definition.definition <logger_definition_id>
``` 