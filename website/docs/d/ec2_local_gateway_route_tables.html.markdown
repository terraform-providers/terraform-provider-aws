---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_route_tables"
description: |-
    Provides information for multiple EC2 Local Gateway Route Tables
---

# Data Source: aws_ec2_local_gateway_route_tables

Provides information for multiple EC2 Local Gateway Route Tables, such as their identifiers.

## Example Usage

The following shows outputing all Local Gateway Route Table Ids.

```hcl
data "aws_ec2_local_gateway_route_table" "foo" {}

output "foo" {
  value = data.aws_ec2_local_gateway_route_table.foo.ids
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired local gateway route table.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGatewayRouteTables.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Local Gateway Route Table will be selected if any one of the given values matches.

## Attributes Reference

* `id` - AWS Region.
* `ids` - Set of Local Gateway Route Table identifiers
