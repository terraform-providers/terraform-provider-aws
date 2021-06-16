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

```terraform
resource "aws_workspaces_ip_group" "contractors" {
  name        = "Contractors"
  description = "Contractors IP access control group"
  rules {
    source      = "150.24.14.0/24"
    description = "NY"
  }
  rules {
    source      = "125.191.14.85/32"
    description = "LA"
  }
  rules {
    source      = "44.98.100.0/24"
    description = "STL"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the IP group.
* `description` - (Optional) The description of the IP group.
* `rules` - (Optional) One or more pairs specifying the IP group rule (in CIDR format) from which web requests originate.
* `tags` – (Optional) A map of tags assigned to the WorkSpaces directory. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Nested Blocks

### `rules`

#### Arguments

* `source` - (Required) The IP address range, in CIDR notation, e.g. `10.0.0.0/16`
* `description` - (Optional) The description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The IP group identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

WorkSpaces IP groups can be imported using their GroupID, e.g.

```
$ terraform import aws_workspaces_ip_group.example wsipg-488lrtl3k
```

