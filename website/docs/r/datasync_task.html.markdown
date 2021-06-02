---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_task"
description: |-
  Manages an AWS DataSync Task
---

# Resource: aws_datasync_task

Manages an AWS DataSync Task, which represents a configuration for synchronization. Starting an execution of these DataSync Tasks (actually synchronizing files) is performed outside of this Terraform resource.

## Example Usage

```terraform
resource "aws_datasync_task" "example" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = "example"
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    bytes_per_second = -1
  }
}
```

## Example Usage with Scheduling

```terraform
resource "aws_datasync_task" "example" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = "example"
  source_location_arn      = aws_datasync_location_nfs.source.arn

  schedule {
    schedule_expression = "cron(0 12 ? * SUN,WED *)"
  }
}
```

## Example Usage with Filtering

```hcl
resource "aws_datasync_task" "example" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = "example"
  source_location_arn      = aws_datasync_location_nfs.source.arn

  excludes {
    filter_type = "SIMPLE_PATTERN"
    value       = "/folder1|/folder2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `destination_location_arn` - (Required) Amazon Resource Name (ARN) of destination DataSync Location.
* `source_location_arn` - (Required) Amazon Resource Name (ARN) of source DataSync Location.
* `cloudwatch_log_group_arn` - (Optional) Amazon Resource Name (ARN) of the CloudWatch Log Group that is used to monitor and log events in the sync task.
* `excludes` - (Optional) Filter rules that determines which files to exclude from a task.
* `name` - (Optional) Name of the DataSync Task.
* `options` - (Optional) Configuration block containing option that controls the default behavior when you start an execution of this DataSync Task. For each individual task execution, you can override these options by specifying an overriding configuration in those executions.
* `schedule` - (Optional) Specifies a schedule used to periodically transfer files from a source to a destination location.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Task. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### options Argument Reference

~> **NOTE:** If `atime` is set to `BEST_EFFORT`, `mtime` must be set to `PRESERVE`. If `atime` is set to `NONE`, `mtime` must be set to `NONE`.

The following arguments are supported inside the `options` configuration block:

* `atime` - (Optional) A file metadata that shows the last time a file was accessed (that is when the file was read or written to). If set to `BEST_EFFORT`, the DataSync Task attempts to preserve the original (that is, the version before sync `PREPARING` phase) `atime` attribute on all source files. Valid values: `BEST_EFFORT`, `NONE`. Default: `BEST_EFFORT`.
* `bytes_per_second` - (Optional) Limits the bandwidth utilized. For example, to set a maximum of 1 MB, set this value to `1048576`. Value values: `-1` or greater. Default: `-1` (unlimited).
* `gid` - (Optional) Group identifier of the file's owners. Valid values: `BOTH`, `INT_VALUE`, `NAME`, `NONE`. Default: `INT_VALUE` (preserve integer value of the ID).
* `log_level` - (Optional) Determines the type of logs that DataSync publishes to a log stream in the Amazon CloudWatch log group that you provide. Valid values: `OFF`, `BASIC`, `TRANSFER`. Default: `OFF`.
* `mtime` - (Optional) A file metadata that indicates the last time a file was modified (written to) before the sync `PREPARING` phase. Value values: `NONE`, `PRESERVE`. Default: `PRESERVE`.
* `overwrite_mode` - (Optional) Determines whether files at the destination should be overwritten or preserved when copying files. Valid values: `ALWAYS`, `NEVER`. Default: `ALWAYS`.
* `posix_permissions` - (Optional) Determines which users or groups can access a file for a specific purpose such as reading, writing, or execution of the file. Valid values: `NONE`, `PRESERVE`. Default: `PRESERVE`.
* `preserve_deleted_files` - (Optional) Whether files deleted in the source should be removed or preserved in the destination file system. Valid values: `PRESERVE`, `REMOVE`. Default: `PRESERVE`.
* `preserve_devices` - (Optional) Whether the DataSync Task should preserve the metadata of block and character devices in the source files system, and recreate the files with that device name and metadata on the destination. The DataSync Task can’t sync the actual contents of such devices, because many of the devices are non-terminal and don’t return an end of file (EOF) marker. Valid values: `NONE`, `PRESERVE`. Default: `NONE` (ignore special devices).
* `task_queueing` - (Optional) Determines whether tasks should be queued before executing the tasks. Valid values: `ENABLED`, `DISABLED`. Default `ENABLED`.
* `transfer_mode` - (Optional) Determines whether DataSync transfers only the data and metadata that differ between the source and the destination location, or whether DataSync transfers all the content from the source, without comparing to the destination location. Valid values: `CHANGED`, `ALL`. Default: `CHANGED`
* `uid` - (Optional) User identifier of the file's owners. Valid values: `BOTH`, `INT_VALUE`, `NAME`, `NONE`. Default: `INT_VALUE` (preserve integer value of the ID).
* `verify_mode` - (Optional) Whether a data integrity verification should be performed at the end of a task execution after all data and metadata have been transferred. Valid values: `NONE`, `POINT_IN_TIME_CONSISTENT`, `ONLY_FILES_TRANSFERRED`. Default: `POINT_IN_TIME_CONSISTENT`.

### Schedule

* `schedule_expression` - (Required) Specifies the schedule you want your task to use for repeated executions. For more information, see [Schedule Expressions for Rules](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html).

### excludes Argument Reference

* `filter_type` - (Optional) The type of filter rule to apply. Valid values: `SIMPLE_PATTERN`.
* `value` - (Optional) A single filter string that consists of the patterns to include or exclude. The patterns are delimited by "|" (that is, a pipe), for example: `/folder1|/folder2`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Task.
* `arn` - Amazon Resource Name (ARN) of the DataSync Task.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_datasync_task` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `5m`) How long to wait for DataSync Task availability.

## Import

`aws_datasync_task` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_task.example arn:aws:datasync:us-east-1:123456789012:task/task-12345678901234567
```
