---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway"
description: |-
    Provides details about an EC2 Local Gateway
---

# Data Source: aws_ec2_local_gateway

Provides details about an EC2 Local Gateway.

## Example Usage

The following example shows how one might accept a local gateway id as a variable.

```hcl
variable "local_gateway_id" {}

data "aws_ec2_local_gateway" "selected" {
  id = var.local_gateway_id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Local Gateways in the current region. The given filters must match exactly one
Local Gateway whose data will be exported as attributes.

* `filter` - (Optional) Custom filter block as described below.

* `id` - (Optional) The id of the specific Local Gateway to retrieve.

* `state` - (Optional) The current state of the desired Local Gateway.
  Can be either `"pending"` or `"available"`.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired Local Gateway.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A Local Gateway will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` blocks are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Local Gateway.

The following attributes are additionally exported:

* `outpost_arn` - Amazon Resource Name (ARN) of Outpost
* `owner_id` - AWS account identifier that owns the Local Gateway.
* `state` - State of the local gateway.
