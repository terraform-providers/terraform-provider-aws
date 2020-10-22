---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_access_point"
description: |-
  Provides an Elastic File System (EFS) access point.
---

# Resource: aws_efs_access_point

Provides an Elastic File System (EFS) access point.

## Example Usage

```hcl
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Required) The ID of the file system for which the access point is intended.
* `posix_user` - (Optional) The operating system user and group applied to all file system requests made using the access point. See [Posix User](#posix-user) below.
* `root_directory`- (Optional) Specifies the directory on the Amazon EFS file system that the access point provides access to. See [Root Directory](#root-directory) below.
* `tags` - (Optional) Key-value mapping of resource tags.

### Posix User

The `posix_user` block supports the following:

* `gid` - (Required) The POSIX group ID used for all file system operations using this access point.
* `uid` - (Required) The POSIX user ID used for all file system operations using this access point.
* `secondary_gids` - (Optional) Secondary POSIX group IDs used for all file system operations using this access point.

### Root Directory

The access point exposes the specified file system path as the root directory of your file system to applications using the access point.
NFS clients using the access point can only access data in the access point's RootDirectory and it's subdirectories.

The `root_directory` block supports the following:

* `path` - (Optional) Specifies the path on the EFS file system to expose as the root directory to NFS clients using the access point to access the EFS file system. A path can have up to four subdirectories. If the specified path does not exist, you are required to provide `creation_info`.
* `creation_info` - (Optional) Specifies the POSIX IDs and permissions to apply to the access point's Root Directory. See [Creation Info](#creation-info) below.

### Creation Info

If the `path` specified does not exist, EFS creates the root directory using the `creation_info` settings when a client connects to an access point.

The `creation_info` block supports the following:

* `owner_gid` - (Required) Specifies the POSIX group ID to apply to the `root_directory`.
* `owner_uid` - (Required) Specifies the POSIX user ID to apply to the `root_directory`.
* `permissions` - (Required) Specifies the POSIX permissions to apply to the RootDirectory, in the format of an octal number representing the file's mode bits.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the access point.
* `arn` - Amazon Resource Name of the access point.
* `file_system_arn` - Amazon Resource Name of the file system.

## Import

The EFS access points can be imported using the `id`, e.g.

```
$ terraform import aws_efs_access_point.test fsap-52a643fb
```
