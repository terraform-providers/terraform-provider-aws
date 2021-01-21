---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route"
description: |-
  Manages an EC2 Local Gateway Route
---

# Resource: aws_ec2_local_gateway_route

Manages an EC2 Local Gateway Route. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-networking-components.html#routing).

## Example Usage

```hcl
resource "aws_ec2_local_gateway_route" "example" {
  destination_cidr_block                   = "172.16.0.0/16"
  local_gateway_route_table_id             = data.aws_ec2_local_gateway_route_table.example.id
  local_gateway_virtual_interface_group_id = data.aws_ec2_local_gateway_virtual_interface_group.example.id
}
```

## Argument Reference

The following arguments are required:

* `destination_cidr_block` - (Required) IPv4 CIDR range used for destination matches. Routing decisions are based on the most specific match.
* `local_gateway_route_table_id` - (Required) Identifier of EC2 Local Gateway Route Table.
* `local_gateway_virtual_interface_group_id` - (Required) Identifier of EC2 Local Gateway Virtual Interface Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Local Gateway Route Table identifier and destination CIDR block separated by underscores (`_`)

## Import

`aws_ec2_local_gateway_route` can be imported by using the EC2 Local Gateway Route Table identifier and destination CIDR block separated by underscores (`_`), e.g.

```
$ terraform import aws_ec2_local_gateway_route.example lgw-rtb-12345678_172.16.0.0/16
```
