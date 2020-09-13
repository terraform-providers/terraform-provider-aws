---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_protection"
description: |-
  Enables AWS Shield Advanced for a specific AWS resource.
---

# Resource: aws_shield_protection

Enables AWS Shield Advanced for a specific AWS resource.
The resource can be an Amazon CloudFront distribution, Elastic Load Balancing load balancer, AWS Global Accelerator accelerator, Elastic IP Address, or an Amazon Route 53 hosted zone.

## Example Usage

### Create protection

```hcl
data "aws_availability_zones" "available" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_eip" "example" {
  vpc = true
}

resource "aws_shield_protection" "example" {
  name         = "example"
  resource_arn = "arn:aws:ec2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name for the Protection you are creating.
* `resource_arn` - (Required) The ARN (Amazon Resource Name) of the resource to be protected.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) for the Protection object that is created.

## Import

Shield protection resources can be imported by specifying their ID e.g.

```
$ terraform import aws_shield_protection.example ff9592dc-22f3-4e88-afa1-7b29fde9669a
```
