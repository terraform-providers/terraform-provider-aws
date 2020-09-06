---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost_instance_type"
description: |-
  Information about single Outpost Instance Type.
---

# Data Source: aws_outposts_outpost_instance_type

Information about single Outpost Instance Type.

## Example Usage

```hcl
data "aws_outposts_outpost_instance_type" "example" {
  arn                      = data.aws_outposts_outpost.example.arn
  preferred_instance_types = ["m5.large", "m5.4xlarge"]
}

resource "aws_ec2_instance" "example" {
  # ... other configuration ...

  instance_type = data.aws_outposts_outpost_instance_type.example.instance_type
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost Amazon Resource Name (ARN).

The following arguments are optional:

* `instance_type` - (Optional) Desired instance type. Conflicts with `preferred_instance_types`.
* `preferred_instance_types` - (Optional) Ordered list of preferred instance types. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. Conflicts with `instance_type`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Outpost identifier.
