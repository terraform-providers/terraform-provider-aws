---
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table_association_table_association"
sidebar_current: "docs-aws-resource-ec2-transit-gateway-route-table-association"
description: |-
  Manages an EC2 Transit Gateway Route Table association
---

# aws_ec2_transit_gateway_route_table_association

Manages an EC2 Transit Gateway Route Table association.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_route_table_association" "example" {
  transit_gateway_attachment_id  = "${aws_ec2_transit_gateway_vpc_attachment.example.id}"
  transit_gateway_route_table_id = "${aws_ec2_transit_gateway_route_table.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_attachment_id` - (Required) Identifier of EC2 Transit Gateway Attachment.
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Route Table identifier combined with EC2 Transit Gateway Attachment identifier
* `resource_id` - Identifier of the resource
* `resource_type` - Type of the resource

## Import

`aws_ec2_transit_gateway_route_table_association` can be imported by using the EC2 Transit Gateway Route Table identifier, an underscore, and the EC2 Transit Gateway Attachment identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_route_table_association.example tgw-rtb-12345678_tgw-attach-87654321
```
