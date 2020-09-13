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
* `subnet_ids` - The identifiers of the subnets where the directory resides.
* `tags` – A map of tags assigned to the WorkSpaces directory.
* `workspace_security_group_id` - The identifier of the security group that is assigned to new WorkSpaces.
* `iam_role_id` - The identifier of the IAM role. This is the role that allows Amazon WorkSpaces to make calls to other services, such as Amazon EC2, on your behalf.
* `registration_code` - The registration code for the directory. This is the code that users enter in their Amazon WorkSpaces client application to connect to the directory.
* `directory_name` - The name of the directory.
* `directory_type` - The directory type.
* `customer_user_name` - The user name for the service account.
* `alias` - The directory alias.
* `ip_group_ids` - The identifiers of the IP access control groups associated with the directory.
* `dns_ip_addresses` - The IP addresses of the DNS servers for the directory.
* `self_service_permissions` – The permissions to enable or disable self-service capabilities.
    * `change_compute_type` – Whether WorkSpaces directory users can change the compute type (bundle) for their workspace.
    * `increase_volume_size` – Whether WorkSpaces directory users can increase the volume size of the drives on their workspace.
    * `rebuild_workspace` – Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state.
    * `restart_workspace` – Whether WorkSpaces directory users can restart their workspace.
    * `switch_running_mode` – Whether WorkSpaces directory users can switch the running mode of their workspace.
