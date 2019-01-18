---
layout: "aws"
page_title: "AWS: aws_ssm_maintenance_window"
sidebar_current: "docs-aws-resource-ssm-maintenance-window"
description: |-
  Provides an SSM Maintenance Window resource
---

# aws_ssm_maintenance_window

Provides an SSM Maintenance Window resource

## Example Usage

```hcl
resource "aws_ssm_maintenance_window" "production" {
  name     = "maintenance-window-application"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the maintenance window.
* `schedule` - (Required) The schedule of the Maintenance Window in the form of a [cron](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-maintenance-cron.html) or rate expression.
* `cutoff` - (Required) The number of hours before the end of the Maintenance Window that Systems Manager stops scheduling new tasks for execution.
* `duration` - (Required) The duration of the Maintenance Window in hours.
* `allow_unassociated_targets` - (Optional) Whether targets must be registered with the Maintenance Window before tasks can be defined for those targets.
* `enabled` - (Optional) Whether the maintenance window is enabled. Default: `true`.
* `end_date` - (Optional) Timestamp in [ISO-8601 extended format](https://www.iso.org/iso-8601-date-and-time-format.html) when to no longer run the maintenance window.
* `schedule_timezone` - (Optional) Timezone for schedule in [Internet Assigned Numbers Authority (IANA) Time Zone Database format](https://www.iana.org/time-zones). For example: `America/Los_Angeles`, `etc/UTC`, or `Asia/Seoul`.
* `start_date` - (Optional) Timestamp in [ISO-8601 extended format](https://www.iso.org/iso-8601-date-and-time-format.html) when to begin the maintenance window.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the maintenance window.

## Import
SSM  Maintenance Windows can be imported using the `maintenance window id`, e.g.
```
$ terraform import aws_ssm_maintenance_window.imported-window mw-0123456789
```
