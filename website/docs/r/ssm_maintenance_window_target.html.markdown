---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_maintenance_window_target"
description: |-
  Provides an SSM Maintenance Window Target resource
---

# Resource: aws_ssm_maintenance_window_target

Provides an SSM Maintenance Window Target resource

## Instance Target Example Usage

```hcl
resource "aws_ssm_maintenance_window" "window" {
  name     = "maintenance-window-webapp"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target1" {
  window_id     = aws_ssm_maintenance_window.window.id
  name          = "maintenance-window-target"
  description   = "This is a maintenance window target"
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }
}
```

## Resource Group Target Example Usage

```hcl
resource "aws_ssm_maintenance_window" "window" {
  name     = "maintenance-window-webapp"
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "target1" {
  window_id     = aws_ssm_maintenance_window.window.id
  name          = "maintenance-window-target"
  description   = "This is a maintenance window target"
  resource_type = "RESOURCE_GROUP"

  targets {
    key    = "resource-groups:ResourceTypeFilters"
    values = ["AWS::EC2::Instance"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `window_id` - (Required) The Id of the maintenance window to register the target with.
* `name` - (Optional) The name of the maintenance window target.
* `description` - (Optional) The description of the maintenance window target.
* `resource_type` - (Required) The type of target being registered with the Maintenance Window. Possible values are `INSTANCE` and `RESOURCE_GROUP`.
* `targets` - (Required) The targets to register with the maintenance window. In other words, the instances to run commands on when the maintenance window runs. You can specify targets using instance IDs, resource group names, or tags that have been applied to instances. For more information about these examples formats see
 (https://docs.aws.amazon.com/systems-manager/latest/userguide/mw-cli-tutorial-targets-examples.html)
* `owner_information` - (Optional) User-provided value that will be included in any CloudWatch events raised while running tasks for these targets in this Maintenance Window.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the maintenance window target.

## Import

SSM Maintenance Window targets can be imported using `WINDOW_ID/WINDOW_TARGET_ID`, e.g.

```
$ terraform import aws_ssm_maintenance_window_target.example mw-0c50858d01EXAMPLE/23639a0b-ddbc-4bca-9e72-78d96EXAMPLE
```
