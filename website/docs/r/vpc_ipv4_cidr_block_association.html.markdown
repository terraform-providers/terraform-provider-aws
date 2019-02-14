---
layout: "aws"
page_title: "AWS: aws_vpc_ipv4_cidr_block_association"
sidebar_current: "docs-aws-resource-vpc-ipv4-cidr-block-association"
description: |-
  Associate additional IPv4 CIDR blocks with a VPC
---

# aws_vpc_ipv4_cidr_block_association

Provides a resource to associate additional IPv4 CIDR blocks with a VPC.

When a VPC is created, a primary IPv4 CIDR block for the VPC must be specified.
The `aws_vpc_ipv4_cidr_block_association` resource allows further IPv4 CIDR blocks to be added to the VPC.

## Example Usage

```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "172.2.0.0/16"
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Required) The additional IPv4 CIDR block to associate with the VPC.
* `vpc_id` - (Required) The ID of the VPC to make the association with.

## Timeouts

`aws_vpc_ipv4_cidr_block_association` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating the association
- `delete` - (Default `10 minutes`) Used for destroying the association

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC CIDR association

## Import

`aws_vpc_ipv4_cidr_block_association` can be imported by using the VPC CIDR Association ID, e.g.

```
$ terraform import aws_vpc_ipv4_cidr_block_association.example vpc-cidr-assoc-xxxxxxxx
```
