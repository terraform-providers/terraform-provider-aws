---
layout: "aws"
page_title: "AWS: aws_route"
sidebar_current: "docs-aws-resource-route|"
description: |-
  Provides a resource to create a routing entry in a VPC routing table.
---

# aws_route

Provides a resource to create a routing table entry (a route) in a VPC routing table.

~> **NOTE on Route Tables and Routes:** Terraform currently
provides both a standalone Route resource and a [Route Table](route_table.html) resource with routes
defined in-line. At this time you cannot use a Route Table with in-line routes
in conjunction with any Route resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

## Example usage:

```hcl
resource "aws_route" "r" {
  route_table_id            = "rtb-4fbb3ac4"
  destination_cidr_block    = "10.0.1.0/22"
  vpc_peering_connection_id = "pcx-45ff3dc1"
  depends_on                = ["aws_route_table.testing"]
}
```

##Example IPv6 Usage:

```hcl
resource "aws_vpc" "vpc" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_egress_only_internet_gateway" "egress" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_route" "r" {
  route_table_id              = "rtb-4fbb3ac4"
  destination_ipv6_cidr_block = "::/0"
  egress_only_gateway_id      = "${aws_egress_only_internet_gateway.egress.id}"
}
```

## Argument Reference

The following arguments are supported:

* `route_table_id` - (Required) The ID of the routing table.

One of the following destination arguments must be supplied:

* `destination_cidr_block` - (Optional) The destination CIDR block.
* `destination_ipv6_cidr_block` - (Optional) The destination IPv6 CIDR block.

One of the following target arguments must be supplied:

* `egress_only_gateway_id` - (Optional) Identifier of a VPC Egress Only Internet Gateway.
* `gateway_id` - (Optional) Identifier of a VPC internet gateway or a virtual private gateway.
* `instance_id` - (Optional) Identifier of an EC2 instance.
* `nat_gateway_id` - (Optional) Identifier of a VPC NAT gateway.
* `network_interface_id` - (Optional) Identifier of an EC2 network interface.
* `transit_gateway_id` - (Optional) Identifier of an EC2 Transit Gateway.
* `vpc_peering_connection_id` - (Optional) Identifier of a VPC peering connection.

Note that the default route, mapping the VPC's CIDR block to "local", is
created implicitly and cannot be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

~> **NOTE:** Only the arguments that are configured (one of the above)
will be exported as an attribute once the resource is created.

* `id` - Route Table identifier and destination

## Timeouts

`aws_route` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `2 minutes`) Used for route creation
- `delete` - (Default `5 minutes`) Used for route deletion

## Import

Individual routes can be imported using `ROUTETABLEID_DESTINATION`.

For example, import a route in route table `rtb-656C65616E6F72` with an IPv4 destination CIDR of `10.42.0.0/16` like this:

```console
$ terraform import aws_route.my_route rtb-656C65616E6F72_10.42.0.0/16
```

Import a route in route table `rtb-656C65616E6F72` with an IPv6 destination CIDR of `2620:0:2d0:200::8/125` similarly:

```console
$ terraform import aws_route.my_route rtb-656C65616E6F72_2620:0:2d0:200::8/125
```
