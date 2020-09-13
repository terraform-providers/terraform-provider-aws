---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb"
description: |-
  Provides a Load Balancer data source.
---

# Data Source: aws_lb

~> **Note:** `aws_alb` is known as `aws_lb`. The functionality is identical.

Provides information about a Load Balancer.

This data source can prove useful when a module accepts an LB as an input
variable and needs to, for example, determine the security groups associated
with it, etc.

## Example Usage

```hcl
variable "lb_arn" {
  type    = string
  default = ""
}

variable "lb_name" {
  type    = string
  default = ""
}

data "aws_lb" "test" {
  arn  = var.lb_arn
  name = var.lb_name
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Optional) The full ARN of the load balancer.
* `name` - (Optional) The unique name of the load balancer.

~> **NOTE**: When both `arn` and `name` are specified, `arn` takes precedence.

## Attributes Reference

See the [LB Resource](/docs/providers/aws/r/lb.html) for details on the
returned attributes - they are identical.
