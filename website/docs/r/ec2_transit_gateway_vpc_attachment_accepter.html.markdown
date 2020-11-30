---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_vpc_attachment_accepter"
description: |-
  Manages the accepter's side of an EC2 Transit Gateway VPC Attachment
---

# Resource: aws_ec2_transit_gateway_vpc_attachment_accepter

Manages the accepter's side of an EC2 Transit Gateway VPC Attachment.

When a cross-account (requester's AWS account differs from the accepter's AWS account) EC2 Transit Gateway VPC Attachment
is created, an EC2 Transit Gateway VPC Attachment resource is automatically created in the accepter's account.
The requester can use the `aws_ec2_transit_gateway_vpc_attachment` resource to manage its side of the connection
and the accepter can use the `aws_ec2_transit_gateway_vpc_attachment_accepter` resource to "adopt" its side of the
connection into management.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id

  tags = {
    Name = "Example cross-account attachment"
  }
}
```

A full example of how to how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a VPC in the second account to the Transit Gateway via the `aws_ec2_transit_gateway_vpc_attachment` and `aws_ec2_transit_gateway_vpc_attachment_accepter` resources can be found in [the `./examples/transit-gateway-cross-account-vpc-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/master/examples/transit-gateway-cross-account-vpc-attachment).

## Argument Reference

The following arguments are supported:

* `transit_gateway_attachment_id` - (Required) The ID of the EC2 Transit Gateway Attachment to manage.
* `transit_gateway_default_route_table_association` - (Optional) Boolean whether the VPC Attachment should be associated with the EC2 Transit Gateway association default route table. Default value: `true`.
* `transit_gateway_default_route_table_propagation` - (Optional) Boolean whether the VPC Attachment should propagate routes with the EC2 Transit Gateway propagation default route table. Default value: `true`.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway VPC Attachment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `appliance_mode_support` - Whether Appliance Mode support is enabled. Valid values: `disable`, `enable`.
* `dns_support` - Whether DNS support is enabled. Valid values: `disable`, `enable`.
* `ipv6_support` - Whether IPv6 support is enabled. Valid values: `disable`, `enable`.
* `subnet_ids` - Identifiers of EC2 Subnets.
* `transit_gateway_id` - Identifier of EC2 Transit Gateway.
* `vpc_id` - Identifier of EC2 VPC.
* `vpc_owner_id` - Identifier of the AWS account that owns the EC2 VPC.

## Import

`aws_ec2_transit_gateway_vpc_attachment_accepter` can be imported by using the EC2 Transit Gateway Attachment identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_vpc_attachment_accepter.example tgw-attach-12345678
```
