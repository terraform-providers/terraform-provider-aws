---
layout: "aws"
page_title: "AWS: aws_vpcs"
sidebar_current: "docs-aws-datasource-vpcs"
description: |-
    Provides a list of VPC Ids in a region
---

# Data Source: aws_vpcs

This resource can be useful for getting back a list of VPC Ids for a region.

The following example retrieves a list of VPC Ids with a custom tag of `service` set to a value of "production".

## Example Usage

The following shows outputing all VPC Ids.

```hcl
data "aws_vpcs" "foo" {
  tags = {
    service = "production"
  }
}

output "foo" {
  value = "${data.aws_vpcs.foo.ids}"
}
```

An example use case would be interpolate the `aws_vpcs` output into `count` of an aws_flow_log resource.

```hcl
data "aws_vpcs" "foo" {}

resource "aws_flow_log" "test_flow_log" {
  count = "${length(data.aws_vpcs.foo.ids)}"
  # ...
  vpc_id = "${element(data.aws_vpcs.foo.ids, count.index)}"
  # ...
}

output "foo" {
  value = "${data.aws_vpcs.foo.ids}"
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired vpcs.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - A list of all the VPC Ids found. This data source will fail if none are found.
