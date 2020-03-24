---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_directory"
description: |-
  Provides a directory registration in AWS WorkSpaces Service.
---

# Resource: aws_workspaces_directory

Provides a directory registration in AWS WorkSpaces Service

## Example Usage

```hcl
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "private-a" {
  vpc_id            = "${aws_vpc.main.id}"
  availability_zone = "us-east-1a"
  cidr_block        = "10.0.0.0/24"
}

resource "aws_subnet" "private-b" {
  vpc_id            = "${aws_vpc.main.id}"
  availability_zone = "us-east-1b"
  cidr_block        = "10.0.1.0/24"
}

resource "aws_directory_service_directory" "main" {
  name     = "corp.example.com"
  password = "#S1ncerely"
  size     = "Small"
  vpc_settings {
    vpc_id     = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.private-a.id}", "${aws_subnet.private-b.id}"]
  }
}

resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"

  self_service_permissions = {
    increase_volume_size = true
    rebuild_workspace    = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The directory identifier for registration in WorkSpaces service.
* `subnet_ids` - (Optional) The identifiers of the subnets where the directory resides.
* `tags` – (Optional) A mapping of tags assigned to the WorkSpaces directory.
* `self_service_permissions` – (Optional) The permissions to enable or disable self-service capabilities.

`self_service_permissions` supports the following:

* `change_compute_type` – (Optional) Whether WorkSpaces directory users can change the compute type (bundle) for their workspace. Default `false`.
* `increase_volume_size` – (Optional) Whether WorkSpaces directory users can increase the volume size of the drives on their workspace. Default `false`.
* `rebuild_workspace` – (Optional) Whether WorkSpaces directory users can rebuild the operating system of a workspace to its original state. Default `false`.
* `restart_workspace` – (Optional) Whether WorkSpaces directory users can restart their workspace. Default `true`.
* `switch_running_mode` – (Optional) Whether WorkSpaces directory users can switch the running mode of their workspace. Default `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The WorkSpaces directory identifier.

## Import

Workspaces directory can be imported using the directory ID, e.g.

```
$ terraform import aws_workspaces_directory.main d-4444444444
```
