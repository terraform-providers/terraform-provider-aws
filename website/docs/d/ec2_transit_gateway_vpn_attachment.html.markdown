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
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpn_connection_id  = aws_vpn_connection.example.id
}
```

### Filter

```hcl
data "aws_ec2_transit_gateway_vpn_attachment" "test" {
  filter {
    name   = "resource-id"
    values = ["some-resource"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Optional) Identifier of the EC2 Transit Gateway.
* `vpn_connection_id` - (Optional) Identifier of the EC2 VPN Connection.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `tags` - (Optional) A map of tags, each pair of which must exactly match a pair on the desired Transit Gateway VPN Attachment.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribeTransitGatewayAttachments API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayAttachments.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway VPN Attachment identifier
* `tags` - Key-value tags for the EC2 Transit Gateway VPN Attachment
