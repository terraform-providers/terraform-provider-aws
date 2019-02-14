---
layout: "aws"
page_title: "AWS: aws_route_tables"
sidebar_current: "docs-aws-datasource-route-tables"
description: |-
    Get information on Amazon route tables.
---

# Data Source: aws_route_tables

This resource can be useful for getting back a list of route table ids to be referenced elsewhere.

## Example Usage

The following adds a route for a particular cidr block to every (private
kops) route table in a specified vpc to use a particular vpc peering
connection.

```hcl
data "aws_route_tables" "rts" {
  vpc_id = "${var.vpc_id}"

  filter {
    name   = "tag:kubernetes.io/kops/role"
    values = ["private*"]
  }
}

resource "aws_route" "r" {
  count                     = "${length(data.aws_route_tables.rts.ids)}"
  route_table_id            = "${data.aws_route_tables.rts.ids[count.index]}"
  destination_cidr_block    = "10.0.1.0/22"
  vpc_peering_connection_id = "pcx-0e9a7a9ecd137dc54"
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.

* `vpc_id` - (Optional) The VPC ID that you want to filter from.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired route tables.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeRouteTables.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Route Table will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - A list of all the route table ids found. This data source will fail if none are found.
