---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost_instance_types"
description: |-
  Information about Outpost Instance Types.
---

# Data Source: aws_outposts_outpost_instance_types

Information about Outposts Instance Types.

## Example Usage

```hcl
data "aws_outposts_outpost_instance_types" "example" {
  arn = data.aws_outposts_outpost.example.arn
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost Amazon Resource Name (ARN).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `instance_types` - Set of instance types.
