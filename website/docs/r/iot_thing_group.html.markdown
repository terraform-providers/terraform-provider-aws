---
layout: "aws"
page_title: "AWS: aws_iot_thing_group"
description: |-
    Creates and manages an AWS IoT Thing Group.
---

# Resource: aws_iot_thing_group

Creates and manages an AWS IoT Thing Group.

## Example Usage

```hcl
resource "aws_iot_thing_group" "group" {
  name = "thing_group"

  properties {
    description = "test description"
    attributes = {
        "attr1": "val1",
        "attr2": "val2",
    }
    merge = false 
  }
}
```

## Argument Reference

* `name` - (Required, Forces New Resource). The thing group name to create.
* `parent_group_name` - (Optional, Forces New Resource). The name of the parent thing group.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the thing group.
* `properties` - (Optional). The thing group properties.

Arguments of `properties`:
* `attributes` - (Optional) Map. A JSON string containing up to three key-value pair in JSON format. 
* `merge` - Specifies whether the list of attributes provided in the attributes is merged with the attributes stored in the registry, instead of overwriting them. To remove an attribute, call UpdateThing with an empty attribute value. The merge attribute is only valid while updating.
* `description` - (Optional). The thing group description.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - The ARN of the created AWS IoT Thing Group.
* `version` - The version of the thing group.

## Import

IOT Thing Group can be imported using the name, e.g.

```
$ terraform import aws_iot_thing_group.example example
```
