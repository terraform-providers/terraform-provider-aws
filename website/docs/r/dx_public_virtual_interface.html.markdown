---
layout: "aws"
page_title: "AWS: aws_dx_public_virtual_interface"
sidebar_current: "docs-aws-resource-dx-public-virtual-interface"
description: |-
  Provides a Direct Connect public virtual interface resource.
---

# aws_dx_public_virtual_interface

Provides a Direct Connect public virtual interface resource.

## Example Usage

```hcl
resource "aws_dx_public_virtual_interface" "foo" {
  connection_id = "dxcon-zzzzzzzz"

  name           = "vif-foo"
  vlan           = 4094
  address_family = "ipv4"
  bgp_asn        = 65352

  customer_address = "175.45.176.1/30"
  amazon_address   = "175.45.176.2/30"

  route_filter_prefixes = [
    "210.52.109.0/24",
    "175.45.176.0/22",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `address_family` - (Required) The address family for the BGP peer. `ipv4 ` or `ipv6`.
* `bgp_asn` - (Required) The autonomous system (AS) number for Border Gateway Protocol (BGP) configuration.
* `connection_id` - (Required) The ID of the Direct Connect connection (or LAG) on which to create the virtual interface.
* `name` - (Required) The name for the virtual interface.
* `vlan` - (Required) The VLAN ID.
* `amazon_address` - (Optional) The IPv4 CIDR address to use to send traffic to Amazon. Required for IPv4 BGP peers.
* `bgp_auth_key` - (Optional) The authentication key for BGP configuration.
* `customer_address` - (Optional) The IPv4 CIDR destination address to which Amazon should send traffic. Required for IPv4 BGP peers.
* `route_filter_prefixes` - (Required) A list of routes to be advertised to the AWS network in this region.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual interface.
* `arn` - The ARN of the virtual interface.

## Timeouts

`aws_dx_public_virtual_interface` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating virtual interface
- `delete` - (Default `10 minutes`) Used for destroying virtual interface

## Import

Direct Connect public virtual interfaces can be imported using the `vif id`, e.g.

```
$ terraform import aws_dx_public_virtual_interface.test dxvif-33cc44dd
```
