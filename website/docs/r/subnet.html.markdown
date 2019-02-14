---
layout: "aws"
page_title: "AWS: aws_subnet"
sidebar_current: "docs-aws-resource-subnet"
description: |-
  Provides an VPC subnet resource.
---

# aws_subnet

Provides an VPC subnet resource.

## Example Usage

### Basic Usage

```hcl
resource "aws_subnet" "main" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "Main"
  }
}
```

### Subnets In Secondary VPC CIDR Blocks

When managing subnets in one of a VPC's secondary CIDR blocks created using a [`aws_vpc_ipv4_cidr_block_association`](vpc_ipv4_cidr_block_association.html)
resource, it is recommended to reference that resource's `vpc_id` attribute to ensure correct dependency ordering.

```hcl
resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = "${aws_vpc.main.id}"
  cidr_block = "172.2.0.0/16"
}

resource "aws_subnet" "in_secondary_cidr" {
  vpc_id     = "${aws_vpc_ipv4_cidr_block_association.secondary_cidr.vpc_id}"
  cidr_block = "172.2.0.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) The AZ for the subnet.
* `availability_zone_id` - (Optional) The AZ ID of the subnet.
* `cidr_block` - (Required) The CIDR block for the subnet.
* `ipv6_cidr_block` - (Optional) The IPv6 network range for the subnet,
    in CIDR notation. The subnet size must use a /64 prefix length.
* `map_public_ip_on_launch` -  (Optional) Specify true to indicate
    that instances launched into the subnet should be assigned
    a public IP address. Default is `false`.
* `assign_ipv6_address_on_creation` - (Optional) Specify true to indicate
    that network interfaces created in the specified subnet should be
    assigned an IPv6 address. Default is `false`
* `vpc_id` - (Required) The VPC ID.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the subnet
* `arn` - The ARN of the subnet.
* `ipv6_cidr_block_association_id` - The association ID for the IPv6 CIDR block.
* `owner_id` - The ID of the AWS account that owns the subnet.

## Import

Subnets can be imported using the `subnet id`, e.g.

```
$ terraform import aws_subnet.public_subnet subnet-9d4a7b6c
```
