---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_workspace"
description: |-
  Provides a workspaces in AWS Workspaces Service.
---

# Resource: aws_workspaces_workspace

Provides a workspace in [AWS Workspaces](https://docs.aws.amazon.com/workspaces/latest/adminguide/amazon-workspaces.html) Service

~> **NOTE:** AWS WorkSpaces service requires [`workspaces_DefaultRole`](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspaces-access-control.html#create-default-role) IAM role to operate normally.

## Example Usage

```hcl
data "aws_workspaces_bundle" "value_windows_10" {
  bundle_id = "wsb-bh8rsxt14" # Value with Windows 10 (English)
}

resource "aws_workspaces_workspace" "example" {
  directory_id = aws_workspaces_directory.example.id
  bundle_id    = data.aws_workspaces_bundle.value_windows_10.id
  user_name    = "john.doe"

  root_volume_encryption_enabled = true
  user_volume_encryption_enabled = true
  volume_encryption_key          = "alias/aws/workspaces"

  workspace_properties {
    compute_type_name                         = "VALUE"
    user_volume_size_gib                      = 10
    root_volume_size_gib                      = 80
    running_mode                              = "AUTO_STOP"
    running_mode_auto_stop_timeout_in_minutes = 60
  }

  tags = {
    Department = "IT"
  }
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The ID of the directory for the WorkSpace.
* `bundle_id` - (Required) The ID of the bundle for the WorkSpace.
* `user_name` – (Required) The user name of the user for the WorkSpace. This user name must exist in the directory for the WorkSpace.
* `root_volume_encryption_enabled` - (Optional) Indicates whether the data stored on the root volume is encrypted.
* `user_volume_encryption_enabled` – (Optional) Indicates whether the data stored on the user volume is encrypted.
* `volume_encryption_key` – (Optional) The symmetric AWS KMS customer master key (CMK) used to encrypt data stored on your WorkSpace. Amazon WorkSpaces does not support asymmetric CMKs.
* `tags` - (Optional) The tags for the WorkSpace.
* `workspace_properties` – (Optional) The WorkSpace properties.

`workspace_properties` supports the following:

* `compute_type_name` – (Optional) The compute type. For more information, see [Amazon WorkSpaces Bundles](http://aws.amazon.com/workspaces/details/#Amazon_WorkSpaces_Bundles). Valid values are `VALUE`, `STANDARD`, `PERFORMANCE`, `POWER`, `GRAPHICS`, `POWERPRO` and `GRAPHICSPRO`.
* `root_volume_size_gib` – (Optional) The size of the root volume.
* `running_mode` – (Optional) The running mode. For more information, see [Manage the WorkSpace Running Mode](https://docs.aws.amazon.com/workspaces/latest/adminguide/running-mode.html). Valid values are `AUTO_STOP` and `ALWAYS_ON`.
* `running_mode_auto_stop_timeout_in_minutes` – (Optional) The time after a user logs off when WorkSpaces are automatically stopped. Configured in 60-minute intervals.
* `user_volume_size_gib` – (Optional) The size of the user storage.

### Timeouts

`aws_workspaces_workspace` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for WorkSpace creation.
- `update` - (Default `10 minutes`) Used for WorkSpace updating.
- `delete` - (Default `10 minutes`) Used for WorkSpace termination.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspaces ID.
* `ip_address` - The IP address of the WorkSpace.
* `computer_name` - The name of the WorkSpace, as seen by the operating system.
* `state` - The operational state of the WorkSpace.

## Import

Workspaces can be imported using their ID, e.g.

```
$ terraform import aws_workspaces_workspace.example ws-9z9zmbkhv
```

