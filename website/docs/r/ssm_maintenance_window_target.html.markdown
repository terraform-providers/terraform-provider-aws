---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_maintenance_window_target"
description: |-
  Provides an SSM Maintenance Window Target resource
---

# Resource: aws_ssm_maintenance_window_target

Provides an SSM Maintenance Window Target resource

## Example Usage

```hcl
resource "aws_ssm_maintenance_window" "window" {
  name     = "maintenance-window-webapp"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target1" {
  window_id     = "${aws_ssm_maintenance_window.window.id}"
  name          = "maintenance-window-target"
  description   = "This is a maintenance window target"
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `window_id` - (Required) The Id of the maintenance window to register the target with.
* `name` - (Optional) The name of the maintenance window target.
* `description` - (Optional) The description of the maintenance window target.
* `resource_type` - (Required) The type of target being registered with the Maintenance Window. Possible values `INSTANCE`.
* `targets` - (Required) The targets (either instances or tags). Instances are specified using Key=InstanceIds,Values=InstanceId1,InstanceId2. Tags are specified using Key=tag name,Values=tag value.
* `owner_information` - (Optional) User-provided value that will be included in any CloudWatch events raised while running tasks for these targets in this Maintenance Window.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the maintenance window target.
