---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_fsx_windows_file_system"
description: |-
  Manages an FSx Windows Location within AWS DataSync.
---

# Resource: aws_datasync_location_fsx_windows_file_system

Manages an AWS DataSync FSx Windows Location.

## Example Usage

```hcl
resource "aws_datasync_location_fsx_windows_file_system" "example" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.example.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.example.arn]
}
```

## Argument Reference

The following arguments are supported:

* `fsx_filesystem_arn` - (Required) The Amazon Resource Name (ARN) for the FSx for Windows file system.
* `password` - (Required) The password of the user who has the permissions to access files and folders in the FSx for Windows file system.
* `user` - (Required) The user who has the permissions to access files and folders in the FSx for Windows file system.
* `domain` - (Optional) The name of the Windows domain that the FSx for Windows server belongs to.
* `security_group_arns` - (Optional) The Amazon Resource Names (ARNs) of the security groups that are to use to configure the FSx for Windows file system.
* `subdirectory` - (Optional) Subdirectory to perform actions as source or destination.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `uri` - The URL of the FSx for Windows location that was described.
* `creation_time` - The time that the FSx for Windows location was created.

## Import

`aws_datasync_location_fsx_windows_file_system` can be imported by using the `DataSync-ARN#FSx-Windows-ARN`, e.g.

```
$ terraform import aws_datasync_location_fsx_windows_file_system.example arn:aws:datasync:us-west-2:123456789012:location/loc-12345678901234567#arn:aws:fsx:us-west-2:476956259333:file-system/fs-08e04cd442c1bb94a
```
