---
layout: "aws"
page_title: "AWS: aws_route_table"
sidebar_current: "docs-aws-resource-route-table|"
description: |-
  Provides a resource to create a VPC routing table.
---

# aws_route_table

Provides a resource to create a VPC routing table.

~> **NOTE on Route Tables and Routes:** Terraform currently
provides both a standalone [Route resource](route.html) and a Route Table resource with routes
defined in-line. At this time you cannot use a Route Table with in-line routes
in conjunction with any Route resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

~> **NOTE on `gateway_id` and `nat_gateway_id`:** The AWS API is very forgiving with these two
attributes and the `aws_route_table` resource can be created with a NAT ID specified as a Gateway ID attribute.
This _will_ lead to a permanent diff between your configuration and statefile, as the API returns the correct
parameters in the returned route table. If you're experiencing constant diffs in your `aws_route_table` resources,
the first thing to check is whether or not you're specifying a NAT ID instead of a Gateway ID, or vice-versa.

~> **NOTE on `propagating_vgws` and the `aws_vpn_gateway_route_propagation` resource:**
If the `propagating_vgws` argument is present, it's not supported to _also_
define route propagations using `aws_vpn_gateway_route_propagation`, since
this resource will delete any propagating gateways not explicitly listed in
`propagating_vgws`. Omit this argument when defining route propagation using
the separate resource.

## Example usage with tags:

```hcl
resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.default.id}"

  route {
    cidr_block = "10.0.1.0/24"
    gateway_id = "${aws_internet_gateway.main.id}"
  }

  route {
    ipv6_cidr_block        = "::/0"
    egress_only_gateway_id = "${aws_egress_only_internet_gateway.foo.id}"
  }

  tags = {
    Name = "main"
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The VPC ID.
* `route` - (Optional) A list of route objects. Their keys are documented below.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `propagating_vgws` - (Optional) A list of virtual gateways for propagation.

### route Argument Reference

One of the following destination arguments must be supplied:

* `cidr_block` - (Required) The CIDR block of the route.
* `ipv6_cidr_block` - Optional) The Ipv6 CIDR block of the route

One of the following target arguments must be supplied:

* `egress_only_gateway_id` - (Optional) Identifier of a VPC Egress Only Internet Gateway.
* `gateway_id` - (Optional) Identifier of a VPC internet gateway or a virtual private gateway.
* `instance_id` - (Optional) Identifier of an EC2 instance.
* `nat_gateway_id` - (Optional) Identifier of a VPC NAT gateway.
* `network_interface_id` - (Optional) Identifier of an EC2 network interface.
* `transit_gateway_id` - (Optional) Identifier of an EC2 Transit Gateway.
* `vpc_peering_connection_id` - (Optional) Identifier of a VPC peering connection.

Note that the default route, mapping the VPC's CIDR block to "local", is created implicitly and cannot be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:
~> **NOTE:** Only the target that is entered is exported as a readable
attribute once the route resource is created.

* `id` - The ID of the routing table
* `owner_id` - The ID of the AWS account that owns the route table

## Import

~> **NOTE:** Importing this resource currently adds an `aws_route` resource to the state for each route, in addition to adding the `aws_route_table` resource. If you plan to apply the imported state, avoid the deletion of actual routes by not using in-line routes in your configuration and by naming `aws_route` resources after the `aws_route_table`. For example, if your route table is `aws_route_table.rt`, name routes as `aws_route.rt`, `aws_route.rt-1` and so forth. The behavior of adding `aws_route` resources with the `aws_route_table` resource will be removed in the next major version.

Route Tables can be imported using the route table `id`. For example, to import
route table `rtb-4e616f6d69`, use this command:

```
$ terraform import aws_route_table.public_rt rtb-4e616f6d69
```
