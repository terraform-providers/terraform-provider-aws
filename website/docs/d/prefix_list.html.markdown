---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_prefix-list"
description: |-
    Provides details about a specific prefix list
---

# Data Source: aws_prefix_list

`aws_prefix_list` provides details about a specific prefix list (PL)
in the current region.

This can be used both to validate a prefix list given in a variable
and to obtain the CIDR blocks (IP address ranges) for the associated
AWS service. The latter may be useful e.g. for adding network ACL
rules.

## Example Usage

```hcl
resource "aws_vpc_endpoint" "private_s3" {
  vpc_id       = aws_vpc.foo.id
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_prefix_list" "private_s3" {
  prefix_list_id = aws_vpc_endpoint.private_s3.prefix_list_id
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_network_acl_rule" "private_s3" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = data.aws_prefix_list.private_s3.cidr_blocks[0]
  from_port      = 443
  to_port        = 443
}
```

### Filter

```hcl
data "aws_prefix_list" "test" {
  filter {
    name   = "prefix-list-id"
    values = ["pl-68a54001"]
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
prefix lists. The given filters must match exactly one prefix list
whose data will be exported as attributes.

* `prefix_list_id` - (Optional) The ID of the prefix list to select.
* `name` - (Optional) The name of the prefix list to select.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribePrefixLists API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribePrefixLists.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the selected prefix list.
* `name` - The name of the selected prefix list.
* `cidr_blocks` - The list of CIDR blocks for the AWS service associated with the prefix list.
