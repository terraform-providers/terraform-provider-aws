---
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway"
sidebar_current: "docs-aws-resource-ec2-transit-gateway-x"
description: |-
  Manages an EC2 Transit Gateway
---

# aws_ec2_transit_gateway

Manages an EC2 Transit Gateway.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway" "example" {
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `amazon_side_asn` - (Optional) Private Autonomous System Number (ASN) for the Amazon side of a BGP session. The range is `64512` to `65534` for 16-bit ASNs and `4200000000` to `4294967294` for 32-bit ASNs. Default value: `64512`.
* `auto_accept_shared_attachments` - (Optional) Whether resource attachment requests are automatically accepted. Valid values: `disable`, `enable`. Default value: `disable`.
* `default_route_table_association` - (Optional) Whether resource attachments are automatically associated with the default association route table. Valid values: `disable`, `enable`. Default value: `enable`.
* `default_route_table_propagation` - (Optional) Whether resource attachments automatically propagate routes to the default propagation route table. Valid values: `disable`, `enable`. Default value: `enable`.
* `description` - (Optional) Description of the EC2 Transit Gateway.
* `dns_support` - (Optional) Whether DNS support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway.
* `vpn_ecmp_support` - (Optional) Whether VPN Equal Cost Multipath Protocol support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Amazon Resource Name (ARN)
* `association_default_route_table_id` - Identifier of the default association route table
* `id` - EC2 Transit Gateway identifier
* `owner_id` - Identifier of the AWS account that owns the EC2 Transit Gateway
* `propagation_default_route_table_id` - Identifier of the default propagation route table

## Import

`aws_ec2_transit_gateway` can be imported by using the EC2 Transit Gateway identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway.example tgw-12345678
```
