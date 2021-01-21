---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
description: |-
  Provides a WorkSpaces directory in AWS WorkSpaces Service.
---

# Resource: aws_workspaces_directory

Provides a WorkSpaces directory in AWS WorkSpaces Service.

~> **NOTE:** AWS WorkSpaces service requires [`workspaces_DefaultRole`](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspaces-access-control.html#create-default-role) IAM role to operate normally.

## Example Usage

```hcl
resource "aws_workspaces_directory" "example" {
  directory_id = aws_directory_service_directory.example.id
  subnet_ids = [
    aws_subnet.example_c.id,
    aws_subnet.example_d.id
  ]

  tags = {
    Example = true
  }

  self_service_permissions {
    change_compute_type  = true
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = true
    switch_running_mode  = true
  }

  workspace_access_properties {
    device_type_android    = "ALLOW"
    device_type_chromeos   = "ALLOW"
    device_type_ios        = "ALLOW"
    device_type_osx        = "ALLOW"
    device_type_web        = "DENY"
    device_type_windows    = "DENY"
    device_type_zeroclient = "DENY"
  }

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.example.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = true
    user_enabled_as_local_administrator = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.workspaces_default_service_access,
    aws_iam_role_policy_attachment.workspaces_default_self_service_access
  ]
}

resource "aws_directory_service_directory" "example" {
  name     = "corp.example.com"
  password = "#S1ncerely"
  size     = "Small"

  vpc_settings {
    vpc_id = aws_vpc.example.id
    subnet_ids = [
      aws_subnet.example_a.id,
      aws_subnet.example_b.id
    ]
  }
}

data "aws_iam_policy_document" "workspaces" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "workspaces_default" {
  name               = "workspaces_DefaultRole"
  assume_role_policy = data.aws_iam_policy_document.workspaces.json
}

resource "aws_iam_role_policy_attachment" "workspaces_default_service_access" {
  role       = aws_iam_role.workspaces_default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
}

resource "aws_iam_role_policy_attachment" "workspaces_default_self_service_access" {
  role       = aws_iam_role.workspaces_default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example_a" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1a"
  cidr_block        = "10.0.0.0/24"
}

resource "aws_subnet" "example_b" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1b"
  cidr_block        = "10.0.1.0/24"
}
resource "aws_subnet" "example_c" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1c"
  cidr_block        = "10.0.2.0/24"
}

resource "aws_subnet" "example_d" {
  vpc_id            = aws_vpc.example.id
  availability_zone = "us-east-1d"
  cidr_block        = "10.0.3.0/24"
}
```

### IP Groups

```hcl
resource "aws_workspaces_directory" "example" {
  directory_id = aws_directory_service_directory.example.id

  ip_group_ids = [
    aws_workspaces_ip_group.example.id,
  ]
}

resource "aws_workspaces_ip_group" "example" {
  name = "example"
}
```

## Arguments Reference

The following arguments are supported:

* `directory_id` - (Required) The directory identifier for registration in WorkSpaces service.
* `subnet_ids` - (Optional) The identifiers of the subnets where the directory resides.
* `ip_group_ids` - The identifiers of the IP access control groups associated with the directory.
* `tags` – (Optional) A map of tags assigned to the WorkSpaces directory.
* `self_service_permissions` – (Optional) Permissions to enable or disable self-service capabilities. Defined below.
* `workspace_access_properties` – (Optional) Specifies which devices and operating systems users can use to access their WorkSpaces. Defined below.
* `workspace_creation_properties` – (Optional) Default properties that are used for creating WorkSpaces. Defined below.

### self_service_permissions

* `change_compute_type` – (Optional) Whether WorkSpaces directory users can change the compute type (bundle) for their workspace. Default `false`.
* `increase_volume_size` – (Optional) Whether WorkSpaces directory users can increase the volume size of the drives on their workspace. Default `false`.
* `rebuild_workspace` – (Optional) Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state. Default `false`.
* `restart_workspace` – (Optional) Whether WorkSpaces directory users can restart their workspace. Default `true`.
* `switch_running_mode` – (Optional) Whether WorkSpaces directory users can switch the running mode of their workspace. Default `false`.

### workspace_access_properties

* `device_type_android` – (Optional) Indicates whether users can use Android devices to access their WorkSpaces.
* `device_type_chromeos` – (Optional) Indicates whether users can use Chromebooks to access their WorkSpaces.
* `device_type_ios` – (Optional) Indicates whether users can use iOS devices to access their WorkSpaces.
* `device_type_osx` – (Optional) Indicates whether users can use macOS clients to access their WorkSpaces.
* `device_type_web` – (Optional) Indicates whether users can access their WorkSpaces through a web browser.
* `device_type_windows` – (Optional) Indicates whether users can use Windows clients to access their WorkSpaces.
* `device_type_zeroclient` – (Optional) Indicates whether users can use zero client devices to access their WorkSpaces.

### workspace_creation_properties

-> **Note:** Once you specified `custom_security_group_id` or `default_ou`, there is no way to delete these attributes. If you cleanup them from the configuration, they still be present in state.

* `custom_security_group_id` – (Optional) The identifier of your custom security group. Should relate to the same VPC, where workspaces reside in.
* `default_ou` – (Optional) The default organizational unit (OU) for your WorkSpace directories. Should conform `"OU=<value>,DC=<value>,...,DC=<value>"` pattern.
* `enable_internet_access` – (Optional) Indicates whether internet access is enabled for your WorkSpaces.
* `enable_maintenance_mode` – (Optional) Indicates whether maintenance mode is enabled for your WorkSpaces. For more information, see [WorkSpace Maintenance](https://docs.aws.amazon.com/workspaces/latest/adminguide/workspace-maintenance.html)..
* `user_enabled_as_local_administrator` – (Optional) Indicates whether users are local administrators of their WorkSpaces.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The WorkSpaces directory identifier.
* `alias` - The directory alias.
* `customer_user_name` - The user name for the service account.
* `directory_name` - The name of the directory.
* `directory_type` - The directory type.
* `dns_ip_addresses` - The IP addresses of the DNS servers for the directory.
* `iam_role_id` - The identifier of the IAM role. This is the role that allows Amazon WorkSpaces to make calls to other services, such as Amazon EC2, on your behalf.
* `ip_group_ids` - The identifiers of the IP access control groups associated with the directory.
* `registration_code` - The registration code for the directory. This is the code that users enter in their Amazon WorkSpaces client application to connect to the directory.
* `workspace_security_group_id` - The identifier of the security group that is assigned to new WorkSpaces.

## Import

Workspaces directory can be imported using the directory ID, e.g.

```
$ terraform import aws_workspaces_directory.main d-4444444444
```
