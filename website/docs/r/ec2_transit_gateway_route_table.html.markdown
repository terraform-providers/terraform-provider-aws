---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table"
description: |-
  Manages an EC2 Transit Gateway Route Table
---

# Resource: aws_ec2_transit_gateway_route_table

Manages an EC2 Transit Gateway Route Table.

## Example Usage

```terraform
resource "aws_ec2_transit_gateway_route_table" "example" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Route Table. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - EC2 Transit Gateway Route Table Amazon Resource Name (ARN).
* `default_association_route_table` - Boolean whether this is the default association route table for the EC2 Transit Gateway.
* `default_propagation_route_table` - Boolean whether this is the default propagation route table for the EC2 Transit Gateway.
* `id` - EC2 Transit Gateway Route Table identifier
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_ec2_transit_gateway_route_table` can be imported by using the EC2 Transit Gateway Route Table identifier, e.g.

```
$ terraform import aws_ec2_transit_gateway_route_table.example tgw-rtb-12345678
```
