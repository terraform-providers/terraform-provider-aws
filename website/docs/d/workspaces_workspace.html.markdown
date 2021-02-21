---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_workspace"
description: |-
  Get information about a WorkSpace in AWS Workspaces Service.
---

# Resource: aws_workspaces_workspace

Use this data source to get information about a workspace in [AWS Workspaces](https://docs.aws.amazon.com/workspaces/latest/adminguide/amazon-workspaces.html) Service.

## Example Usage

### Filter By Workspace ID

```hcl
data "aws_workspaces_workspace" "example" {
  workspace_id = "ws-cj5xcxsz5"
}
```

### Filter By Directory ID & User Name

```hcl
data "aws_workspaces_workspace" "example" {
  directory_id = "d-9967252f57"
  user_name    = "Example"
}
```

## Argument Reference

The following arguments are supported:

* `bundle_id` - (Optional) The ID of the bundle for the WorkSpace.
* `directory_id` - (Optional) The ID of the directory for the WorkSpace. You have to specify `user_name` along with `directory_id`. You cannot combine this parameter with `workspace_id`.
* `root_volume_encryption_enabled` - (Optional) Indicates whether the data stored on the root volume is encrypted.
* `tags` - (Optional) The tags for the WorkSpace.
* `user_name` – (Optional) The user name of the user for the WorkSpace. This user name must exist in the directory for the WorkSpace. You cannot combine this parameter with `workspace_id`.
* `user_volume_encryption_enabled` – (Optional) Indicates whether the data stored on the user volume is encrypted.
* `volume_encryption_key` – (Optional) The symmetric AWS KMS customer master key (CMK) used to encrypt data stored on your WorkSpace. Amazon WorkSpaces does not support asymmetric CMKs.
* `workspace_id` - (Optional) The ID of the WorkSpace. You cannot combine this parameter with `directory_id`.
* `workspace_properties` – (Optional) The WorkSpace properties.

`workspace_properties` supports the following:

* `compute_type_name` – (Optional) The compute type. For more information, see [Amazon WorkSpaces Bundles](http://aws.amazon.com/workspaces/details/#Amazon_WorkSpaces_Bundles). Valid values are `VALUE`, `STANDARD`, `PERFORMANCE`, `POWER`, `GRAPHICS`, `POWERPRO` and `GRAPHICSPRO`.
* `root_volume_size_gib` – (Optional) The size of the root volume.
* `running_mode` – (Optional) The running mode. For more information, see [Manage the WorkSpace Running Mode](https://docs.aws.amazon.com/workspaces/latest/adminguide/running-mode.html). Valid values are `AUTO_STOP` and `ALWAYS_ON`.
* `running_mode_auto_stop_timeout_in_minutes` – (Optional) The time after a user logs off when WorkSpaces are automatically stopped. Configured in 60-minute intervals.
* `user_volume_size_gib` – (Optional) The size of the user storage.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The workspaces ID.
* `ip_address` - The IP address of the WorkSpace.
* `computer_name` - The name of the WorkSpace, as seen by the operating system.
* `state` - The operational state of the WorkSpace.
