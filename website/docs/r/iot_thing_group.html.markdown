---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_thing_group"
description: |-
    Creates and manages an AWS IoT Thing Group.
---

# Resource: aws_iot_thing_group

Creates and manages an AWS IoT Thing Group.

## Example Usage

```hcl
resource "aws_iot_thing_group" "parent" {
  name = "parent"
}

resource "aws_iot_thing_group" "example" {
  name = "example"

  parent_group_name = "${aws_iot_thing_group.parent.name}"

  properties {
    attributes = {
      One    = "11111"
      Two    = "TwoTwo"
    }
    description = "This is my thing group"
  }

  tags = {
    terraform = "true"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the Thing Group.
* `parent_group_name` - (Optional) The name of the parent Thing Group.
* `properties` - (Optional) The Thing Group properties. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags

## properties Reference

* `attributes` - (Optional) Map of attributes of the Thing Group.
* `description` - (Optional) A description of the Thing Group.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The Thing Group ID.
* `version` - The current version of the Thing Group record in the registry.
* `arn` - The ARN of the Thing Group.

## Import

IoT Things Groups can be imported using the name, e.g.

```
$ terraform import aws_iot_thing_group.example example
```
