---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_group"
description: |-
    Creates and manages an AWS IoT Greengrass Group
---

# Resource: aws_greengrass_group

## Example Usage

```hcl
resource "aws_greengrass_group" "test" {
  name = "test_group"
}
```

## Argument Reference

* `name` - (Required). The name of the group.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the group.
* `group_version` - (Optional). Information about a group version.

The `group_version` object takes such following argument.
* `connector_definition_version_arn` - (Optional) String. The ARN of the connector definition version for this group.
* `core_definition_version_arn` - (Optional) String. The ARN of the core definition version for this group.
* `device_definition_version_arn` - (Optional) String. The ARN of the device definition version for this group.
* `function_definition_version_arn` - (Optional) String. The ARN of the function definition version for this group.
* `logger_definition_version_arn` - (Optional) String. The ARN of the logger definition version for this group.
* `resource_definition_version_arn` - (Optional) String. The ARN of the resource definition version for this group.
* `subscription_definition_version_arn` - (Optional) String. The ARN of the subscription definition version for this group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_id` - The id of the group
* `arn` - The ARN of the group

## Environment variables
If you use `device_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import

IoT Greengrass Groups can be imported using the `name`, e.g.

```
$ terraform import aws_greengrass_group.group <group_id>
```
