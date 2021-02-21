---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_route_table"
description: |-
    Provides details about a specific Route Table
---

# Data Source: aws_route_table

`aws_route_table` provides details about a specific Route Table.

This resource can prove useful when a module accepts a Subnet ID as an input variable and needs to, for example, add a route in the Route Table.

## Example Usage

The following example shows how one might accept a Route Table ID as a variable and use this data source to obtain the data necessary to create a route.

```hcl
variable "subnet_id" {}

data "aws_route_table" "selected" {
  subnet_id = var.subnet_id
}

resource "aws_route" "route" {
  route_table_id            = data.aws_route_table.selected.id
  destination_cidr_block    = "10.0.1.0/22"
  vpc_peering_connection_id = "pcx-45ff3dc1"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available Route Table in the current region. The given filters must match exactly one Route Table whose data will be exported as attributes.

The following arguments are optional:

* `filter` - (Optional) Configuration block. Detailed below.
* `gateway_id` - (Optional) ID of an Internet Gateway or Virtual Private Gateway which is connected to the Route Table (not exported if not passed as a parameter).
* `route_table_id` - (Optional) ID of the specific Route Table to retrieve.
* `subnet_id` - (Optional) ID of a Subnet which is connected to the Route Table (not exported if not passed as a parameter).
* `tags` - (Optional) Map of tags, each pair of which must exactly match a pair on the desired Route Table.
* `vpc_id` - (Optional) ID of the VPC that the desired Route Table belongs to.

### filter

Complex filters can be expressed using one or more `filter` blocks.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeRouteTables.html).
* `values` - (Required) Set of values that are accepted for the given field. A Route Table will be selected if any one of the given values matches.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - ARN of the route table.
* `associations` - List of associations with attributes detailed below.
* `owner_id` - ID of the AWS account that owns the route table.
* `routes` - List of routes with attributes detailed below.

### routes

When relevant, routes are also exported with the following attributes:

For destinations:

* `cidr_block` - CIDR block of the route.
* `ipv6_cidr_block` - IPv6 CIDR block of the route.

For targets:

* `carrier_gateway_id` - ID of the Carrier Gateway.
* `egress_only_gateway_id` - ID of the Egress Only Internet Gateway.
* `gateway_id` - Internet Gateway ID.
* `instance_id` - EC2 instance ID.
* `local_gateway_id` - Local Gateway ID.
* `nat_gateway_id` - NAT Gateway ID.
* `network_interface_id` - ID of the elastic network interface (eni) to use.
* `transit_gateway_id` - EC2 Transit Gateway ID.
* `vpc_endpoint_id` - VPC Endpoint ID.
* `vpc_peering_connection_id` - VPC Peering ID.

### associations

Associations are also exported with the following attributes:

* `gateway_id` - Gateway ID. Only set when associated with an Internet Gateway or Virtual Private Gateway.
* `main` - Whether the association is due to the main route table.
* `route_table_association_id` - Association ID.
* `route_table_id` - Route Table ID.
* `subnet_id` - Subnet ID. Only set when associated with a subnet.
