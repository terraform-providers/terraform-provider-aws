---
layout: "aws"
page_title: "AWS: aws_dx_hosted_private_virtual_interface"
sidebar_current: "docs-aws-resource-dx-hosted-private-virtual-interface"
description: |-
  Provides a Direct Connect hosted private virtual interface resource.
---

# aws_dx_hosted_private_virtual_interface

Provides a Direct Connect hosted private virtual interface resource. This resource represents the allocator's side of the hosted virtual interface.
A hosted virtual interface is a virtual interface that is owned by another AWS account.

## Example Usage

```hcl
resource "aws_dx_hosted_private_virtual_interface" "foo" {
  connection_id = "dxcon-zzzzzzzz"

  name           = "vif-foo"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352
}
```

## Argument Reference

The following arguments are supported:

* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Required) The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `connection_id` - (Required) The ID of the Direct Connect connection (or LAG) on which to create the virtual interface.
* `name` - (Required) The name for the virtual interface.
* `owner_account_id` - (Required) The AWS account that will own the new virtual interface.
* `vlan` - (Required) The VLAN ID.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `mtu` - (Optional) The maximum transmission unit (MTU) is the size, in bytes, of the largest permissible packet that can be passed over the connection. The MTU of a virtual private interface can be either `1500` or `9001` (jumbo frames). Default is `1500`.
* `bgp_auth_key` - (Optional) The authentication key for BGP configuration.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.
* `jumbo_frame_capable` - Indicates whether jumbo frames (9001 MTU) are supported.

## Timeouts

`aws_dx_hosted_private_virtual_interface` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating virtual interface
- `update` - (Default `10 minutes`) Used for virtual interface modifications
- `delete` - (Default `10 minutes`) Used for destroying virtual interface

## Import

Direct Connect hosted private virtual interfaces can be imported using the `vif id`, e.g.

```
$ terraform import aws_dx_hosted_private_virtual_interface.test dxvif-33cc44dd
```
