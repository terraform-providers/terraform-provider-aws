---
layout: "aws"
page_title: "AWS: aws_route"
sidebar_current: "docs-aws-datasource-route"
description: |-
    Provides details about a specific Route
---

# Data Source: aws_route

`aws_route` provides details about a specific Route.

This resource can prove useful when finding the resource
associated with a CIDR. For example, finding the peering
connection associated with a CIDR value.

## Example Usage

The following example shows how one might use a CIDR value to find a network interface id
and use this to create a data source of that network interface.

```hcl
variable "subnet_id" {}

data "aws_route_table" "selected" {
  subnet_id = "${var.subnet_id}"
}

data "aws_route" "route" {
  route_table_id         = "${aws_route_table.selected.id}"
  destination_cidr_block = "10.0.1.0/24"
}

data "aws_network_interface" "interface" {
  network_interface_id = "${data.aws_route.route.network_interface_id}"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Route in the current region. The given filters must match exactly one
Route whose data will be exported as attributes.

* `route_table_id` - (Required) The id of the specific Route Table containing the Route entry.

* `destination_cidr_block` - (Optional) The CIDR block of the Route belonging to the Route Table.

* `destination_ipv6_cidr_block` - (Optional) The IPv6 CIDR block of the Route belonging to the Route Table.

* `egress_only_gateway_id` - (Optional) The Egress Only Gateway ID of the Route belonging to the Route Table.

* `gateway_id` - (Optional) The Gateway ID of the Route belonging to the Route Table.

* `instance_id` - (Optional) The Instance ID of the Route belonging to the Route Table.

* `nat_gateway_id` - (Optional) The NAT Gateway ID of the Route belonging to the Route Table.

* `transit_gateway_id` - (Optional) The EC2 Transit Gateway ID of the Route belonging to the Route Table.

* `vpc_peering_connection_id` - (Optional) The VPC Peering Connection ID of the Route belonging to the Route Table.

* `network_interface_id` - (Optional) The Network Interface ID of the Route belonging to the Route Table.

## Attributes Reference

All of the argument attributes are also exported as
result attributes when there is data available. For example, the `vpc_peering_connection_id` field will be empty when the route is attached to a Network Interface.
