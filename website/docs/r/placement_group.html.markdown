---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_placement_group"
description: |-
  Provides an EC2 placement group.
---

# Resource: aws_placement_group

Provides an EC2 placement group. Read more about placement groups
in [AWS Docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html).

## Example Usage

```hcl
resource "aws_placement_group" "web" {
  name     = "hunky-dory-pg"
  strategy = "cluster"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the placement group.
* `strategy` - (Required) The placement strategy.
* `tags` - (Optional) Key-value mapping of resource tags.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the placement group.
* `placement_group_id` - The ID of the placement group.

## Import

Placement groups can be imported using the `name`, e.g.

```
$ terraform import aws_placement_group.prod_pg production-placement-group
```
