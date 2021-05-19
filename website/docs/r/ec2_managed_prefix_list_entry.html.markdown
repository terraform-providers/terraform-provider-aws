---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_ec2_managed_prefix_list_entry"
description: |-
  Provides a managed prefix list entry resource.
---

# Resource: aws_ec2_managed_prefix_list_entry

Provides a managed prefix list entry resource.

~> **NOTE on Managed Prefix Lists and Managed Prefix List Entries:** Terraform
currently provides both a standalone Managed Prefix List Entry resource (a single entry),
and a [Managed Prefix List resource](ec2_managed_prefix_list.html) with entries defined
in-line. At this time you cannot use a Managed Prefix List with in-line rules in
conjunction with any Managed Prefix List Entry resources. Doing so will cause a conflict
of entries and will overwrite entries.

## Example Usage

Basic usage

```terraform
resource "aws_ec2_managed_prefix_list" "example" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Env = "live"
  }
}

resource "aws_ec2_managed_prefix_list_entry" "entry_1" {
  cidr           = aws_vpc.example.cidr_block
  description    = "Primary"
  prefix_list_id = aws_ec2_managed_prefix_list.entry.id
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Required) CIDR block of this entry.
* `description` - (Optional) Description of this entry. Due to API limitations, updating only the description of an existing entry requires temporarily removing and re-adding the entry.
* `prefix_list_id` - (Required) CIDR block of this entry.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the managed prefix list entry.
* `version` - Latest version of this prefix list entry.

## Import

Prefix Lists can be imported using the `prefix_list_id` and `cidr` separated by a `,`, e.g.

```
$ terraform import aws_ec2_managed_prefix_list_entry.default pl-0570a1d2d725c16be,10.0.3.0/24
```
