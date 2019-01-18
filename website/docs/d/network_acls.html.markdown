---
layout: "aws"
page_title: "AWS: aws_network_acls"
sidebar_current: "docs-aws-datasource-network-acls"
description: |-
    Provides a list of network ACL ids for a VPC
---

# Data Source: aws_network_acls

## Example Usage

The following shows outputing all network ACL ids in a vpc.

```hcl
data "aws_network_acls" "example" {
  vpc_id = "${var.vpc_id}"
}

output "example" {
  value = "${data.aws_network_acls.example.ids}"
}
```

The following example retrieves a list of all network ACL ids in a VPC with a custom
tag of `Tier` set to a value of "Private".

```hcl
data "aws_network_acls" "example" {
  vpc_id = "${var.vpc_id}"

  tags = {
    Tier = "Private"
  }
}
```

The following example retrieves a network ACL id in a VPC which associated
with specific subnet.

```hcl
data "aws_network_acls" "example" {
  vpc_id = "${var.vpc_id}"

  filter {
    name   = "association.subnet-id"
    values = ["${aws_subnet.test.id}"]
  }
}
```

## Argument Reference

* `vpc_id` - (Optional) The VPC ID that you want to filter from.

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the desired network ACLs.

* `filter` - (Optional) Custom filter block as described below.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeNetworkAcls.html).

* `values` - (Required) Set of values that are accepted for the given field.
  A VPC will be selected if any one of the given values matches.

## Attributes Reference

* `ids` - A list of all the network ACL ids found. This data source will fail if none are found.
