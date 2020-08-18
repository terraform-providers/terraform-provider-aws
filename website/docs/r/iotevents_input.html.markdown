---
subcategory: "IoTEvents"
layout: "aws"
page_title: "AWS: aws_iotevents_input"
sidebar_current: "docs-aws-resource-iotevents-input"
description: |-
    Creates and manages an AWS IoTEvents Input
---

# Resource: aws_iotevents_input

## Example Usage

```hcl
resource "aws_iotevents_input" "input_example" {
  name        = "MyInput"
  description = "Example Input"

  definition {
    attribute {
      json_path = "temperature"
    }

    attribute {
      json_path = "humidity"
    }
  }
}
```

## Argument Reference

* `name` - (Required) The name of the input.
* `description` - (Optional) The description of the input.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the input.

The `definition` - (Required) The definition of the input. Object takes the following arguments:

* `attribute` - (Required) The attributes from the JSON payload that are made available by the input. Inputs are derived from messages sent to the AWS IoT Events system using `BatchPutMessage`. Each such message contains a JSON payload, and those attributes (and their paired values) specified here are available for use in the `condition` expressions used by detectors. Object takes the following arguments:
    * `json_path` = (Required) An expression that specifies an attribute-value pair in a JSON structure.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the input
* `arn` - The ARN of the input.

## Import

IoTEvents Input can be imported using the `name`, e.g.

```
$ terraform import aws_iotevents_input.input <name>
```
