---
layout: "aws"
page_title: "AWS: efs_mount_target"
sidebar_current: "docs-aws-datasource-efs-mount-target"
description: |-
  Provides an Elastic File System Mount Target (EFS) data source.
---

# Data Source: aws_efs_mount_target

Provides information about an Elastic File System Mount Target (EFS).

## Example Usage

```hcl
variable "mount_target_id" {
  type    = "string"
  default = ""
}

data "aws_efs_mount_target" "by_id" {
  mount_target_id = "${var.mount_target_id}"
}
```

## Argument Reference

The following arguments are supported:

* `mount_target_id` - (Required) ID of the mount target that you want to have described

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `file_system_arn` - Amazon Resource Name of the file system for which the mount target is intended.
* `file_system_id` - ID of the file system for which the mount target is intended.
* `subnet_id` - ID of the mount target's subnet.
* `ip_address` - Address at which the file system may be mounted via the mount target.
* `security_groups` - List of VPC security group IDs attached to the mount target.
* `dns_name` - The DNS name for the given subnet/AZ per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).
* `network_interface_id` - The ID of the network interface that Amazon EFS created when it created the mount target.

