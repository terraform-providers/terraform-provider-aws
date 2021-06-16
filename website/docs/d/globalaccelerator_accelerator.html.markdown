---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_accelerator"
description: |-
  Provides a Global Accelerator accelerator data source.
---

# Data Source: aws_globalaccelerator_accelerator

Provides information about a Global Accelerator accelerator.

## Example Usage

```terraform
variable "accelerator_arn" {
  type    = string
  default = ""
}

variable "accelerator_name" {
  type    = string
  default = ""
}

data "aws_globalaccelerator_accelerator" "example" {
  arn  = var.accelerator_arn
  name = var.accelerator_name
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Optional) The full ARN of the Global Accelerator.
* `name` - (Optional) The unique name of the Global Accelerator.

~> **NOTE**: When both `arn` and `name` are specified, `arn` takes precedence.

## Attributes Reference

website/docs/r/globalaccelerator_accelerator.markdown
See the [`aws_globalaccelerator_accelerator` resource](/docs/providers/aws/r/globalaccelerator_accelerator.html) for details on the
returned attributes - they are identical.
