---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list"
description: |-
  Provides a managed prefix list resource.
---

# Resource: aws_ec2_managed_prefix_list

Provides a managed prefix list resource.

~> **NOTE on `max_entries`:** When you reference a Prefix List in a resource,
the maximum number of entries for the prefix lists counts as the same number of rules
or entries for the resource. For example, if you create a prefix list with a maximum
of 20 entries and you reference that prefix list in a security group rule, this counts
as 20 rules for the security group.

## Example Usage

Basic usage

```hcl
resource "aws_ec2_managed_prefix_list" "example" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr        = aws_vpc.example.cidr_block
    description = "Primary"
  }

  entry {
    cidr        = aws_vpc_ipv4_cidr_block_association.example.cidr_block
    description = "Secondary"
  }

  tags = {
    Env = "live"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of this resource. The name must not start with `com.amazonaws`.
* `address_family` - (Required, Forces new resource) The address family (`IPv4` or `IPv6`) of
    this prefix list.
* `entry` - (Optional) Can be specified multiple times for each prefix list entry.
    Each entry block supports fields documented below. Different entries may have
    overlapping CIDR blocks, but a particular CIDR should not be duplicated.
* `max_entries` - (Required, Forces new resource) The maximum number of entries that
    this prefix list can contain.
* `tags` - (Optional) A map of tags to assign to this resource.

The `entry` block supports:

* `cidr` - (Required) The CIDR block of this entry.
* `description` - (Optional) Description of this entry.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the prefix list.
* `arn` - The ARN of the prefix list.
* `owner_id` - The ID of the AWS account that owns this prefix list.
* `version` - The latest version of this prefix list.

## Import

Prefix Lists can be imported using the `id`, e.g.

```
$ terraform import aws_ec2_managed_prefix_list.default pl-0570a1d2d725c16be
```
