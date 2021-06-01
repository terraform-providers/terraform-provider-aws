---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_thing_type"
description: |-
    Creates and manages an AWS IoT Thing Type.
---

# Resource: aws_iot_thing_type

Creates and manages an AWS IoT Thing Type.

## Example Usage

```hcl
resource "aws_iot_thing_type" "foo" {
  name = "my_iot_thing"
}
```

## Argument Reference

* `name` - (Required, Forces New Resource) The name of the thing type.
* `deprecated` - (Optional, Defaults to false) Whether the thing type is deprecated. If true, no new things could be associated with this type.
* `properties` - (Optional), Configuration block that can contain the following properties of the thing type:
    * `description` - (Optional, Forces New Resource) The description of the thing type.
    * `searchable_attributes` - (Optional, Forces New Resource) A list of searchable thing attribute names.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the created AWS IoT Thing Type.

## Import

IOT Thing Types can be imported using the name, e.g.

```
$ terraform import aws_iot_thing_type.example example
```
