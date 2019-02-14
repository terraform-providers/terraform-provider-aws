---
layout: "aws"
page_title: "AWS: aws_iot_thing"
sidebar_current: "docs-aws-resource-iot-thing"
description: |-
    Creates and manages an AWS IoT Thing.
---

# aws_iot_thing

Creates and manages an AWS IoT Thing.

## Example Usage

```hcl
resource "aws_iot_thing" "example" {
  name = "example"

  attributes = {
    First = "examplevalue"
  }
}
```

## Argument Reference

* `name` - (Required) The name of the thing.
* `attributes` - (Optional) Map of attributes of the thing.
* `thing_type_name` - (Optional) The thing type name.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `default_client_id` - The default client ID.
* `version` - The current version of the thing record in the registry.
* `arn` - The ARN of the thing.

## Import

IOT Things can be imported using the name, e.g.

```
$ terraform import aws_iot_thing.example example
```
