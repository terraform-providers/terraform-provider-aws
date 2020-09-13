---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_egress_only_internet_gateway"
description: |-
  Provides a resource to create an egress-only Internet gateway.
---

# Resource: aws_egress_only_internet_gateway

[IPv6 only] Creates an egress-only Internet gateway for your VPC.
An egress-only Internet gateway is used to enable outbound communication
over IPv6 from instances in your VPC to the Internet, and prevents hosts
outside of your VPC from initiating an IPv6 connection with your instance.

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_egress_only_internet_gateway" "example" {
  vpc_id = aws_vpc.example.id

  tags = {
    Name = "main"
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The VPC ID to create in.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the egress-only Internet gateway.

## Import

Egress-only Internet gateways can be imported using the `id`, e.g.

```
$ terraform import aws_egress_only_internet_gateway.example eigw-015e0e244e24dfe8a
```
