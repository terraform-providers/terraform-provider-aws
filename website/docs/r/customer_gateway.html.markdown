---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_customer_gateway"
description: |-
  Provides a customer gateway inside a VPC. These objects can be
  connected to VPN gateways via VPN connections, and allow you to
  establish tunnels between your network and the VPC.
---

# Resource: aws_customer_gateway

Provides a customer gateway inside a VPC. These objects can be connected to VPN gateways via VPN connections, and allow you to establish tunnels between your network and the VPC.

## Example Usage

```hcl
resource "aws_customer_gateway" "main" {
  bgp_asn    = 65000
  ip_address = "172.83.124.10"
  type       = "ipsec.1"

  tags = {
    Name = "main-customer-gateway"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bgp_asn` - (Required) The gateway's Border Gateway Protocol (BGP) Autonomous System Number (ASN).
* `device_name` - (Optional) A name for the customer gateway device.
* `ip_address` - (Required) The IP address of the gateway's Internet-routable external interface.
* `type` - (Required) The type of customer gateway. The only type AWS
  supports at this time is "ipsec.1".
* `tags` - (Optional) Tags to apply to the gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The amazon-assigned ID of the gateway.
* `arn` - The ARN of the customer gateway.
* `bgp_asn` - The gateway's Border Gateway Protocol (BGP) Autonomous System Number (ASN).
* `device_name` - A name for the customer gateway device.
* `ip_address` - The IP address of the gateway's Internet-routable external interface.
* `type` - The type of customer gateway.
* `tags` - Tags applied to the gateway.


## Import

Customer Gateways can be imported using the `id`, e.g.

```
$ terraform import aws_customer_gateway.main cgw-b4dc3961
```
