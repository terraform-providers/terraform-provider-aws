---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list"
description: |-
    Provides details about a specific managed prefix list
---

# Data Source: aws_ec2_managed_prefix_list

`aws_ec2_managed_prefix_list` provides details about a specific AWS prefix list or
customer-managed prefix list in the current region.

## Example Usage

### Find the regional DynamoDB prefix list

```hcl
data "aws_region" "current" {}

data "aws_ec2_managed_prefix_list" "example" {
  name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"
}
```

### Find a managed prefix list using filters

```hcl
data "aws_ec2_managed_prefix_list" "example" {
  filter {
    name   = "prefix-list-name"
    values = ["my-prefix-list"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
prefix lists. The given filters must match exactly one prefix list
whose data will be exported as attributes.

* `id` - (Optional) The ID of the prefix list to select.
* `name` - (Optional) The name of the prefix list to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the EC2 [DescribeManagedPrefixLists](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeManagedPrefixLists.html) API Reference.
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the selected prefix list.
* `arn` - The ARN of the selected prefix list.
* `name` - The name of the selected prefix list.
* `entries` - The set of entries in this prefix list. Each entry is an object with `cidr` and `description`.
* `owner_id` - The Account ID of the owner of a customer-managed prefix list, or `AWS` otherwise.
* `address_family` - The address family of the prefix list. Valid values are `IPv4` and `IPv6`.
* `max_entries` - When then prefix list is managed, the maximum number of entries it supports, or null otherwise.
* `tags` - A map of tags assigned to the resource.
