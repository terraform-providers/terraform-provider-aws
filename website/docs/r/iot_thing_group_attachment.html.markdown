---
layout: "aws"
page_title: "AWS: aws_iot_thing_group_attachment"
description: |-
    Allow to add IoT Thing to IoT Thing Group.
---

# Resource: aws_iot_thing_group_attachment

Allow to add IoT Thing to IoT Thing Group.

## Example Usage

```hcl
resource "aws_iot_thing_group_attachment" "test_attachment" {
	thing_name = "test_thing_name"
	thing_group_name = "test_thing_group_name"
	override_dynamics_group = false
}
```

## Argument Reference

* `thing_name` - (Required, Forces New Resource). The name of the thing to add to a group.
* `thing_group_name` - (Required, Forces New Resource). The name of the group to which you are adding a thing.
* `override_dynamics_group` - (Optional) Bool. Override dynamic thing groups with static thing groups when 10-group limit is reached. If a thing belongs to 10 thing groups, and one or more of those groups are dynamic thing groups, adding a thing to a static group removes the thing from the last dynamic group.

## Import

IOT Thing Group Attachment can be imported using the name of thing and thing group.

```
$ terraform import aws_iot_thing_group_attachment.test_attachment thing_name/thing_group
```
