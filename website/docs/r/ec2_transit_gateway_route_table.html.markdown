---
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table"
sidebar_current: "docs-aws-resource-ec2-transit-gateway-route-table-x"
description: |-
  Manages an EC2 Transit Gateway Route Table
---

# aws_ec2_transit_gateway_route_table

Manages an EC2 Transit Gateway Route Table.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_route_table" "example" {
  transit_gateway_id = "${aws_ec2_transit_gateway.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Route Table.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `default_association_route_table` - Boolean whether this is the default association route table for the EC2 Transit Gateway.
* `default_propagation_route_table` - Boolean whether this is the default propagation route table for the EC2 Transit Gateway.
* `id` - EC2 Transit Gateway Route Table identifier

## Import

`aws_ec2_transit_gateway_route_table` can be imported by using the EC2 Transit Gateway Route Table identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_route_table.example tgw-rtb-12345678
```
