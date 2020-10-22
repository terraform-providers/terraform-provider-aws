---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
description: |-
  Retrieve information about an AWS WorkSpaces directory.
---

# Data Source: aws_workspaces_directory

Retrieve information about an AWS WorkSpaces directory.

## Example Usage

```hcl
data "aws_workspaces_directory" "example" {
  directory_id = "d-9067783251"
}
```

## Argument Reference

* `directory_id` - (Required) The directory identifier for registration in WorkSpaces service.

## Attributes Reference

* `id` - The WorkSpaces directory identifier.
* `alias` - The directory alias.
* `customer_user_name` - The user name for the service account.
* `directory_name` - The name of the directory.
* `directory_type` - The directory type.
* `dns_ip_addresses` - The IP addresses of the DNS servers for the directory.
* `iam_role_id` - The identifier of the IAM role. This is the role that allows Amazon WorkSpaces to make calls to other services, such as Amazon EC2, on your behalf.
* `ip_group_ids` - The identifiers of the IP access control groups associated with the directory.
* `registration_code` - The registration code for the directory. This is the code that users enter in their Amazon WorkSpaces client application to connect to the directory.
* `self_service_permissions` – The permissions to enable or disable self-service capabilities.
* `subnet_ids` - The identifiers of the subnets where the directory resides.
* `tags` – A map of tags assigned to the WorkSpaces directory.
* `workspace_creation_properties` – The default properties that are used for creating WorkSpaces. Defined below.
* `workspace_security_group_id` - The identifier of the security group that is assigned to new WorkSpaces. Defined below.

### self_service_permissions

* `change_compute_type` – Whether WorkSpaces directory users can change the compute type (bundle) for their workspace.
* `increase_volume_size` – Whether WorkSpaces directory users can increase the volume size of the drives on their workspace.
* `rebuild_workspace` – Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state.
* `restart_workspace` – Whether WorkSpaces directory users can restart their workspace.
* `switch_running_mode` – Whether WorkSpaces directory users can switch the running mode of their workspace.

### workspace_creation_properties

* `custom_security_group_id` – The identifier of your custom security group. Should relate to the same VPC, where workspaces reside in.
* `default_ou` – The default organizational unit (OU) for your WorkSpace directories.
* `enable_internet_access` – Indicates whether internet access is enabled for your WorkSpaces.
* `enable_maintenance_mode` – Indicates whether maintenance mode is enabled for your WorkSpaces. For more information, see [WorkSpace Maintenance](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspace-maintenance.html).
* `user_enabled_as_local_administrator` – Indicates whether users are local administrators of their WorkSpaces.
