---
layout: "aws"
page_title: "AWS: aws_datasync_task"
sidebar_current: "docs-aws-resource-datasync-task"
description: |-
  Manages an AWS DataSync Task
---

# aws_datasync_task

Manages an AWS DataSync Task, which represents a configuration for synchronization. Starting an execution of these DataSync Tasks (actually synchronizing files) is performed outside of this Terraform resource.

## Example Usage

```hcl
resource "aws_datasync_task" "example" {
  destination_location_arn = "${aws_datasync_location_s3.destination.arn}"
  name                     = "example"
  source_location_arn      = "${aws_datasync_location_nfs.source.arn}"

  options {
    bytes_per_second = -1
  }
}
```

## Argument Reference

The following arguments are supported:

* `destination_location_arn` - (Required) Amazon Resource Name (ARN) of destination DataSync Location.
* `source_location_arn` - (Required) Amazon Resource Name (ARN) of source DataSync Location.
* `cloudwatch_log_group_arn` - (Optional) Amazon Resource Name (ARN) of the CloudWatch Log Group that is used to monitor and log events in the sync task.
* `name` - (Optional) Name of the DataSync Task.
* `options` - (Optional) Configuration block containing option that controls the default behavior when you start an execution of this DataSync Task. For each individual task execution, you can override these options by specifying an overriding configuration in those executions.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Task.

### options Argument Reference

~> **NOTE:** If `atime` is set to `BEST_EFFORT`, `mtime` must be set to `PRESERVE`. If `atime` is set to `NONE`, `mtime` must be set to `NONE`.

The following arguments are supported inside the `options` configuration block:

* `atime` - (Optional) A file metadata that shows the last time a file was accessed (that is when the file was read or written to). If set to `BEST_EFFORT`, the DataSync Task attempts to preserve the original (that is, the version before sync `PREPARING` phase) `atime` attribute on all source files. Valid values: `BEST_EFFORT`, `NONE`. Default: `BEST_EFFORT`.
* `bytes_per_second` - (Optional) Limits the bandwidth utilized. For example, to set a maximum of 1 MB, set this value to `1048576`. Value values: `-1` or greater. Default: `-1` (unlimited).
* `gid` - (Optional) Group identifier of the file's owners. Valid values: `BOTH`, `INT_VALUE`, `NAME`, `NONE`. Default: `INT_VALUE` (preserve integer value of the ID).
* `mtime` - (Optional) A file metadata that indicates the last time a file was modified (written to) before the sync `PREPARING` phase. Value values: `NONE`, `PRESERVE`. Default: `PRESERVE`.
* `posix_permissions` - (Optional) Determines which users or groups can access a file for a specific purpose such as reading, writing, or execution of the file. Valid values: `BEST_EFFORT`, `NONE`, `PRESERVE`. Default: `PRESERVE`.
* `preserve_deleted_files` - (Optional) Whether files deleted in the source should be removed or preserved in the destination file system. Valid values: `PRESERVE`, `REMOVE`. Default: `PRESERVE`.
* `preserve_devices` - (Optional) Whether the DataSync Task should preserve the metadata of block and character devices in the source files system, and recreate the files with that device name and metadata on the destination. The DataSync Task can’t sync the actual contents of such devices, because many of the devices are non-terminal and don’t return an end of file (EOF) marker. Valid values: `NONE`, `PRESERVE`. Default: `NONE` (ignore special devices).
* `uid` - (Optional) User identifier of the file's owners. Valid values: `BOTH`, `INT_VALUE`, `NAME`, `NONE`. Default: `INT_VALUE` (preserve integer value of the ID).
* `verify_mode` - (Optional) Whether a data integrity verification should be performed at the end of a task execution after all data and metadata have been transferred. Valid values: `NONE`, `POINT_IN_TIME_CONSISTENT`. Default: `POINT_IN_TIME_CONSISTENT`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Task.
* `arn` - Amazon Resource Name (ARN) of the DataSync Task.

## Timeouts

`aws_datasync_task` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `5m`) How long to wait for DataSync Task availability.

## Import

`aws_datasync_task` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_task.example arn:aws:datasync:us-east-1:123456789012:task/task-12345678901234567
```
