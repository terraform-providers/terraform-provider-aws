---
layout: "aws"
page_title: "AWS: aws_internet_gateway"
sidebar_current: "docs-aws-datasource-internet_gateway"
description: |-
    Provides details about a specific Internet Gateway
---

# Data Source: aws_internet_gateway

`aws_internet_gateway` provides details about a specific Internet Gateway.

## Example Usage

```hcl
variable "vpc_id" {}

data "aws_internet_gateway" "default" {
  filter {
    name   = "attachment.vpc-id"
    values = ["${var.vpc_id}"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Internet Gateway in the current region. The given filters must match exactly one
Internet Gateway whose data will be exported as attributes.

* `internet_gateway_id` - (Optional) The id of the specific Internet Gateway to retrieve.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired Internet Gateway.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html).

* `values` - (Required) Set of values that are accepted for the given field.
  An Internet Gateway will be selected if any one of the given values matches.

## Attributes Reference

All of the argument attributes except `filter` block are also exported as
result attributes. This data source will complete the data by populating
any fields that are not included in the configuration with the data for
the selected Internet Gateway.

`attachments` are also exported with the following attributes, when there are relevants:
Each attachement supports the following:

* `owner_id` - The ID of the AWS account that owns the internet gateway.
* `state` - The current state of the attachment between the gateway and the VPC. Present only if a VPC is attached
* `vpc_id` - The ID of an attached VPC.
