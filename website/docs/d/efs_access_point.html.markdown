---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_access_point"
description: |-
  Provides an Elastic File System (EFS) Access Point data source.
---

# Data Source: aws_efs_access_point

Provides information about an Elastic File System (EFS) Access Point.

## Example Usage

```hcl
data "aws_efs_access_point" "test" {
  access_point_id = "fsap-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `access_point_id` - (Required) The ID that identifies the file system.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the access point.
* `arn` - Amazon Resource Name of the file system.
* `file_system_arn` - Amazon Resource Name of the file system.
* `file_system_id` - The ID of the file system for which the access point is intended.
* `posix_user` - Single element list containing operating system user and group applied to all file system requests made using the access point.
    * `gid` - Group ID
    * `secondary_gids` - Secondary group IDs
    * `uid` - User Id
* `root_directory`- Single element list containing information on the directory on the Amazon EFS file system that the access point provides access to.
    * `creation_info` - Single element list containing information on the creation permissions of the directory
        * `owner_gid` - POSIX owner group ID
        * `owner_uid` - POSIX owner user ID
        * `permissions` - POSIX permissions mode
    * `path` - Path exposed as the root directory
* `tags` - Key-value mapping of resource tags.
