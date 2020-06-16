---
layout: "aws"
page_title: "AWS: aws_dx_private_virtual_interface"
sidebar_current: "docs-aws-datasource-dx-private-virtual-interface"
description: |-
  Retrieve information about a Direct Connect private virtual interface resource.
---

# Data Source: aws_dx_private_virtual_interface

Retrieve information about a Direct Connect private virtual interface resource.

## Example Usage

```hcl
data "aws_dx_private_virtual_interface" "example" {
	virtual_interface_id    = "${aws_dx_private_virtual_interface.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `virtual_interface_id` - (Required) The ID of the virtual interface.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the virtual interface.
* `jumbo_frame_capable` - Indicates whether jumbo frames (9001 MTU) are supported.
* `address_family` -  The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `connection_id` - The ID of the Direct Connect connection (or LAG) on which to create the virtual interface.
* `name` - The name for the virtual interface.
* `vlan` - The VLAN ID.
* `amazon_address` - The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `mtu` - The maximum transmission unit (MTU) is the size, in bytes, of the largest permissible packet that can be passed over the connection.
The MTU of a virtual private interface can be either `1500` or `9001` (jumbo frames). Default is `1500`.
* `bgp_auth_key` - The authentication key for BGP configuration.
* `customer_address` - The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.
* `dx_gateway_id` - The ID of the Direct Connect gateway to which to connect the virtual interface.
* `tags` - A mapping of tags to assign to the resource.
* `vpn_gateway_id` - The ID of the [virtual private gateway](vpn_gateway.html) to which to connect the virtual interface.
