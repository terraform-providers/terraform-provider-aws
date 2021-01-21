---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_ip_group"
description: |-
  Provides an IP access control group in AWS WorkSpaces Service.
---

# Resource: aws_workspaces_ip_group

Provides an IP access control group in AWS WorkSpaces Service

## Example Usage

```hcl
resource "aws_workspaces_ip_group" "contractors" {
  name        = "Contractors"
  description = "Contractors IP access control group"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the IP group.
* `description` - (Optional) The description of the IP group.
* `rules` - (Optional) One or more pairs specifying the IP group rule (in CIDR format) from which web requests originate.

## Nested Blocks

### `rules`

#### Arguments

* `source` - (Required) The IP address range, in CIDR notation, e.g. `10.0.0.0/16`
* `description` - (Optional) The description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The IP group identifier.

## Import

WorkSpaces IP groups can be imported using their GroupID, e.g.

```
$ terraform import aws_workspaces_ip_group.example wsipg-488lrtl3k
```

