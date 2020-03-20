---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpn_attachment"
description: |-
  Get information on an EC2 Transit Gateway VPN Attachment
---

# Data Source: aws_ec2_transit_gateway_vpn_attachment

Get information on an EC2 Transit Gateway VPN Attachment.

## Example Usage

### By Transit Gateway and VPN Connection Identifiers

```hcl
data "aws_ec2_transit_gateway_vpn_attachment" "example" {
  transit_gateway_id = "${aws_ec2_transit_gateway.example.id}"
  vpn_connection_id  = "${aws_vpn_connection.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required) Identifier of the EC2 Transit Gateway.
* `vpn_connection_id` - (Required) Identifier of the EC2 VPN Connection.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway VPN Attachment identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPN Attachment
