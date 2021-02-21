---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_prefix_list_reference"
description: |-
  Manages an EC2 Transit Gateway Prefix List Reference
---

# Resource: aws_ec2_transit_gateway_prefix_list_reference

Manages an EC2 Transit Gateway Prefix List Reference.

## Example Usage

### Attachment Routing

```hcl
resource "aws_ec2_transit_gateway_prefix_list_reference" "example" {
  prefix_list_id                 = aws_ec2_managed_prefix_list.example.id
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.example.association_default_route_table_id
}
```

### Blackhole Routing

```hcl
resource "aws_ec2_transit_gateway_prefix_list_reference" "example" {
  blackhole                      = true
  prefix_list_id                 = aws_ec2_managed_prefix_list.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.example.association_default_route_table_id
}
```

## Argument Reference

The following arguments are required:

* `prefix_list_id` - (Required) Identifier of EC2 Prefix List.
* `transit_gateway_route_table_id` - (Required) Identifier of EC2 Transit Gateway Route Table.

The following arguments are optional:

* `blackhole` - (Optional) Indicates whether to drop traffic that matches the Prefix List. Defaults to `false`.
* `transit_gateway_attachment_id` - (Optional) Identifier of EC2 Transit Gateway Attachment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Route Table identifier and EC2 Prefix List identifier, separated by an underscore (`_`)

## Import

`aws_ec2_transit_gateway_prefix_list_reference` can be imported by using the EC2 Transit Gateway Route Table identifier and EC2 Prefix List identifier, separated by an underscore (`_`), e.g.

```console
$ terraform import aws_ec2_transit_gateway_prefix_list_reference.example tgw-rtb-12345678_pl-12345678
```
