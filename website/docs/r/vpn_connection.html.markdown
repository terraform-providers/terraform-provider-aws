---
layout: "aws"
page_title: "AWS: aws_vpn_connection"
sidebar_current: "docs-aws-resource-vpn-connection"
description: |-
  Manages an EC2 VPN connection. These objects can be connected to customer gateways, and allow you to establish tunnels between your network and Amazon.
---

# aws_vpn_connection

Manages an EC2 VPN connection. These objects can be connected to customer gateways, and allow you to establish tunnels between your network and Amazon.

~> **Note:** All arguments including `tunnel1_preshared_key` and `tunnel2_preshared_key` will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

~> **Note:** The CIDR blocks in the arguments `tunnel1_inside_cidr` and `tunnel2_inside_cidr` must have a prefix of /30 and be a part of a specific range.
[Read more about this in the AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VpnTunnelOptionsSpecification.html).

## Example Usage

### EC2 Transit Gateway

```hcl
resource "aws_ec2_transit_gateway" "example" {}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = "${aws_customer_gateway.example.id}"
  transit_gateway_id  = "${aws_ec2_transit_gateway.example.id}"
  type                = "${aws_customer_gateway.example.type}"
}
```

### Virtual Private Gateway

```hcl
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpn_gateway" "vpn_gateway" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "main" {
  vpn_gateway_id      = "${aws_vpn_gateway.vpn_gateway.id}"
  customer_gateway_id = "${aws_customer_gateway.customer_gateway.id}"
  type                = "ipsec.1"
  static_routes_only  = true
}
```

## Argument Reference

The following arguments are required:

* `customer_gateway_id` - (Required) The ID of the customer gateway.
* `type` - (Required) The type of VPN connection. The only type AWS supports at this time is "ipsec.1".

One of the following arguments is required:

* `transit_gateway_id` - (Optional) The ID of the EC2 Transit Gateway.
* `vpn_gateway_id` - (Optional) The ID of the Virtual Private Gateway.

Other arguments:

* `static_routes_only` - (Optional, Default `false`) Whether the VPN connection uses static routes exclusively. Static routes must be used for devices that don't support BGP.
* `tags` - (Optional) Tags to apply to the connection.
* `tunnel1_inside_cidr` - (Optional) The CIDR block of the inside IP addresses for the first VPN tunnel.
* `tunnel2_inside_cidr` - (Optional) The CIDR block of the second IP addresses for the first VPN tunnel.
* `tunnel1_preshared_key` - (Optional) The preshared key of the first VPN tunnel.
* `tunnel2_preshared_key` - (Optional) The preshared key of the second VPN tunnel.

~> **Note:** The preshared key must be between 8 and 64 characters in length and cannot start with zero(0). Allowed characters are alphanumeric characters, periods(.) and underscores(_).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The amazon-assigned ID of the VPN connection.
* `customer_gateway_configuration` - The configuration information for the VPN connection's customer gateway (in the native XML format).
* `customer_gateway_id` - The ID of the customer gateway to which the connection is attached.
* `static_routes_only` - Whether the VPN connection uses static routes exclusively.
* `tags` - Tags applied to the connection.
* `tunnel1_address` - The public IP address of the first VPN tunnel.
* `tunnel1_cgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (Customer Gateway Side).
* `tunnel1_vgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (VPN Gateway Side).
* `tunnel1_preshared_key` - The preshared key of the first VPN tunnel.
* `tunnel1_bgp_asn` - The bgp asn number of the first VPN tunnel.
* `tunnel1_bgp_holdtime` - The bgp holdtime of the first VPN tunnel.
* `tunnel2_address` - The public IP address of the second VPN tunnel.
* `tunnel2_cgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (Customer Gateway Side).
* `tunnel2_vgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (VPN Gateway Side).
* `tunnel2_preshared_key` - The preshared key of the second VPN tunnel.
* `tunnel2_bgp_asn` - The bgp asn number of the second VPN tunnel.
* `tunnel2_bgp_holdtime` - The bgp holdtime of the second VPN tunnel.
* `type` - The type of VPN connection.
* `vpn_gateway_id` - The ID of the virtual private gateway to which the connection is attached.


## Import

VPN Connections can be imported using the `vpn connection id`, e.g.

```
$ terraform import aws_vpn_connection.testvpnconnection vpn-40f41529
```
