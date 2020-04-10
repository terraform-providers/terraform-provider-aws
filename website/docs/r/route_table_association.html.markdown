---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_route_table_association"
description: |-
  Provides a resource to create an association between a route table and a subnet or a route table and an internet gateway or virtual private gateway.
---

# Resource: aws_route_table_association

Provides a resource to create an association between a route table and a subnet or a route table and an
internet gateway or virtual private gateway.

## Example Usage

```hcl
resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.foo.id
  route_table_id = aws_route_table.bar.id
}
```

```hcl
resource "aws_route_table_association" "b" {
  gateway_id     = aws_internet_gateway.foo.id
  route_table_id = aws_route_table.bar.id
}
```

## Argument Reference

~> **NOTE:** Please note that one of either `subnet_id` or `gateway_id` is required.

The following arguments are supported:

* `subnet_id` - (Optional) The subnet ID to create an association. Conflicts with `gateway_id`.
* `gateway_id` - (Optional) The gateway ID to create an association. Conflicts with `subnet_id`.
* `route_table_id` - (Required) The ID of the routing table to associate with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association

## Import

~> **NOTE:** Attempting to associate a route table with a subnet or gateway, where either
is already associated, will result in an error (e.g.,
`Resource.AlreadyAssociated: the specified association for route table
rtb-4176657279 conflicts with an existing association`) unless you first
import the original association.

EC2 Route Table Associations can be imported using the associated resource ID and Route Table ID
separated by a forward slash (`/`).

For example with EC2 Subnets:

```
$ terraform import aws_route_table_association.assoc subnet-6777656e646f6c796e/rtb-656c65616e6f72
```

For example with EC2 Internet Gateways:

```
$ terraform import aws_route_table_association.assoc igw-01b3a60780f8d034a/rtb-656c65616e6f72
```
