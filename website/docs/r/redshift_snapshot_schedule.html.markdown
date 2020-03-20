---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_snapshot_schedule"
description: |-
  Provides an Redshift Snapshot Schedule resource.
---

# Resource: aws_redshift_snapshot_schedule

## Example Usage

```hcl
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-redshift-snapshot-schedule"
  definitions = [
    "rate(12 hours)",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `identifier` - (Optional, Forces new resource) The snapshot schedule identifier. If omitted, Terraform will assign a random, unique identifier.
* `identifier_prefix` - (Optional, Forces new resource) Creates a unique
identifier beginning with the specified prefix. Conflicts with `identifier`.
* `description` - (Optional) The description of the snapshot schedule.
* `definitions` - (Optional) The definition of the snapshot schedule. The definition is made up of schedule expressions, for example `cron(30 12 *)` or `rate(12 hours)`.
* `force_destroy` - (Optional) Whether to destroy all associated clusters with this snapshot schedule on deletion. Must be enabled and applied before attempting deletion.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Import

Redshift Snapshot Schedule can be imported using the `identifier`, e.g.

```
$ terraform import aws_redshift_snapshot_schedule.default tf-redshift-snapshot-schedule
```
