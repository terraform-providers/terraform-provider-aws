---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment"
description: |-
  Manages an EC2 Transit Gateway VPC Attachment
---

# Resource: aws_ec2_transit_gateway_vpc_attachment

Manages an EC2 Transit Gateway VPC Attachment. For examples of custom route table association and propagation, see the EC2 Transit Gateway Networking Examples Guide.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a VPC in the second account to the Transit Gateway via the `aws_ec2_transit_gateway_vpc_attachment` and `aws_ec2_transit_gateway_vpc_attachment_accepter` resources can be found in [the `./examples/transit-gateway-cross-account-vpc-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/master/examples/transit-gateway-cross-account-vpc-attachment).

## Argument Reference

The following arguments are supported:

* `subnet_ids` - (Required) Identifiers of EC2 Subnets.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `vpc_id` - (Required) Identifier of EC2 VPC.
* `appliance_mode_support` - (Optional) Whether Appliance Mode support is enabled. If enabled, a traffic flow between a source and destination uses the same Availability Zone for the VPC attachment for the lifetime of that flow. Valid values: `disable`, `enable`. Default value: `disable`.
* `dns_support` - (Optional) Whether DNS support is enabled. Valid values: `disable`, `enable`. Default value: `enable`.
* `ipv6_support` - (Optional) Whether IPv6 support is enabled. Valid values: `disable`, `enable`. Default value: `disable`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway VPC Attachment.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the VPC Attachment should be associated with the EC2 Transit Gateway association default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the VPC Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. This cannot be configured or perform drift detection with Resource Access Manager shared EC2 Transit Gateways. Default value: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Import

`aws_ec2_transit_gateway_vpc_attachment` can be imported by using the EC2 Transit Gateway Attachment identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_vpc_attachment.example tgw-attach-12345678
```
