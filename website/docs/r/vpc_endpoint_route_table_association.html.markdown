---
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_route_table_association"
sidebar_current: "docs-aws-resource-vpc-endpoint-route-table-association"
description: |-
  Manages a VPC Endpoint Route Table Association
---

# aws_vpc_endpoint_route_table_association

Manages a VPC Endpoint Route Table Association

## Example Usage

```hcl
resource "aws_vpc_endpoint_route_table_association" "example" {
  route_table_id  = "${aws_route_table.example.id}"
  vpc_endpoint_id = "${aws_vpc_endpoint.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `route_table_id` - (Required) Identifier of the EC2 Route Table to be associated with the VPC Endpoint.
* `vpc_endpoint_id` - (Required) Identifier of the VPC Endpoint with which the EC2 Route Table will be associated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A hash of the EC2 Route Table and VPC Endpoint identifiers.
