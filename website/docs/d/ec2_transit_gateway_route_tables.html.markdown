---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_route_tables"
description: |-
   Provides information for multiple EC2 Transit Gateway Route Tables
---

# Data Source: aws_ec2_transit_gateway_route_tables

Provides information for multiple EC2 Transit Gateway Route Tables, such as their identifiers.

## Example Usage

The following shows outputing all Transit Gateway Route Table Ids.

```terraform
data "aws_ec2_transit_gateway_route_tables" "example" {}

output "example" {
  value = data.aws_ec2_transit_gateway_route_table.example.ids
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired transit gateway route table.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayRouteTables.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Transit Gateway Route Table will be selected if any one of the given values matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `ids` - Set of Transit Gateway Route Table identifiers.
