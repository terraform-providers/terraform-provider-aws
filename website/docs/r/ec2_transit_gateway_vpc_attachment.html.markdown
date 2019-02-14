---
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment"
sidebar_current: "docs-aws-resource-ec2-transit-gateway-vpc-attachment"
description: |-
  Manages an EC2 Transit Gateway VPC Attachment
---

# aws_ec2_transit_gateway_vpc_attachment

Manages an EC2 Transit Gateway VPC Attachment. For examples of custom route table association and propagation, see the EC2 Transit Gateway Networking Examples Guide.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = ["${aws_subnet.example.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.example.id}"
  vpc_id             = "${aws_vpc.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `subnet_ids` - (Required) Identifiers of EC2 Subnets.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `vpc_id` - (Required) Identifier of EC2 VPC.
* `dns_support` - (Optional) Whether DNS support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.
* `ipv6_support` - (Optional) Whether IPv6 support is enabled. Valid values: `disable`, `enable`. Default value: `disable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway VPC Attachment.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the VPC Attachment should be associated with the EC2 Transit Gateway association default route table. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the VPC Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. Default value: `true`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Import

`aws_ec2_transit_gateway_vpc_attachment` can be imported by using the EC2 Transit Gateway Attachment identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_vpc_attachment.example tgw-attach-12345678
```
