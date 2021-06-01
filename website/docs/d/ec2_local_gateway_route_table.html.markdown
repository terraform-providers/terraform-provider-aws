---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route_table"
description: |-
    Provides details about an EC2 Local Gateway Route Table
---

# Data Source: aws_ec2_local_gateway_route_table

Provides details about an EC2 Local Gateway Route Table.

This data source can prove useful when a module accepts a local gateway route table id as
an input variable and needs to, for example, find the associated Outpost or Local Gateway.

## Example Usage

The following example returns a specific local gateway route table ID

```hcl
variable "aws_ec2_local_gateway_route_table" {}

data "aws_ec2_local_gateway_route_table" "selected" {
  local_gateway_route_table_id = var.aws_ec2_local_gateway_route_table
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Local Gateway Route Tables in the current region. The given filters must match exactly one
Local Gateway Route Table whose data will be exported as attributes.

* `local_gateway_route_table_id` - (Optional) Local Gateway Route Table Id assigned to desired local gateway route table

* `local_gateway_id` - (Optional) The id of the specific local gateway route table to retrieve.

* `outpost_arn` - (Optional) The arn of the Outpost the local gateway route table is associated with.

* `state` - (Optional) The state of the local gateway route table.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired local gateway route table.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGatewayRouteTables.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A local gateway route table will be selected if any one of the given values matches.
