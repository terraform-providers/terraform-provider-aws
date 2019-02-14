---
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_table"
sidebar_current: "docs-aws-datasource-ec2-transit-gateway-route-table"
description: |-
  Get information on an EC2 Transit Gateway Route Table
---

# Data Source: aws_ec2_transit_gateway_route_table

Get information on an EC2 Transit Gateway Route Table.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_route_table" "example" {
  filter {
    name   = "default-association-route-table"
    values = ["true"]
  }

  filter {
    name   = "transit-gateway-id"
    values = ["tgw-12345678"]
  }
}
```

### By Identifier

```hcl
data "aws_ec2_transit_gateway_route_table" "example" {
  id = "tgw-rtb-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway Route Table.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `default_association_route_table` - Boolean whether this is the default association route table for the EC2 Transit Gateway
* `default_propagation_route_table` - Boolean whether this is the default propagation route table for the EC2 Transit Gateway
* `id` - EC2 Transit Gateway Route Table identifier
* `transit_gateway_id` - EC2 Transit Gateway identifier
* `tags` - Key-value tags for the EC2 Transit Gateway Route Table
