---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_peering_connections"
description: |-
    Lists peering connections
---

# Data Source: aws_vpc_peering_connections

Use this data source to get IDs of Amazon VPC peering connections
To get more details on each connection, use the data resource [aws_vpc_peering_connection](/docs/providers/aws/d/vpc_peering_connection.html)

Note: To use this data source in a count, the resources should exist before trying to access
the data source, as noted in [issue 4149](https://github.com/hashicorp/terraform/issues/4149)

## Example Usage

```hcl
# Declare the data source
data "aws_vpc_peering_connections" "pcs" {
  filter {
    name   = "requester-vpc-info.vpc-id"
    values = [aws_vpc.foo.id]
  }
}

# get the details of each resource
data "aws_vpc_peering_connection" "pc" {
  count = length(data.aws_vpc_peering_connections.pcs.ids)
  id    = data.aws_vpc_peering_connections.pcs.ids[count.index]
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC peering connections.

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired VPC Peering Connection.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcPeeringConnections.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC Peering Connection will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` are also exported as result attributes.

* `id` - AWS Region.
* `ids` - The IDs of the VPC Peering Connections.
