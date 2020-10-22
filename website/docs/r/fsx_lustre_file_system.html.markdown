---
subcategory: "File System (FSx)"
layout: "aws"
page_title: "AWS: aws_fsx_lustre_file_system"
description: |-
  Manages a FSx Lustre File System.
---

# Resource: aws_fsx_lustre_file_system

Manages a FSx Lustre File System. See the [FSx Lustre Guide](https://docs.aws.amazon.com/fsx/latest/LustreGuide/what-is.html) for more information.

## Example Usage

```hcl
resource "aws_fsx_lustre_file_system" "example" {
  import_path      = "s3://${aws_s3_bucket.example.bucket}"
  storage_capacity = 1200
  subnet_ids       = [aws_subnet.example.id]
}
```

## Argument Reference

The following arguments are supported:

* `storage_capacity` - (Required) The storage capacity (GiB) of the file system. Minimum of `1200`. Storage capacity is provisioned in increments of 3,600 GiB.
* `subnet_ids` - (Required) A list of IDs for the subnets that the file system will be accessible from. File systems currently support only one subnet. The file server is also launched in that subnet's Availability Zone.
* `export_path` - (Optional) S3 URI (with optional prefix) where the root of your Amazon FSx file system is exported. Can only be specified with `import_path` argument and the path must use the same Amazon S3 bucket as specified in `import_path`. Set equal to `import_path` to overwrite files on export. Defaults to `s3://{IMPORT BUCKET}/FSxLustre{CREATION TIMESTAMP}`.
* `import_path` - (Optional) S3 URI (with optional prefix) that you're using as the data repository for your FSx for Lustre file system. For example, `s3://example-bucket/optional-prefix/`.
* `imported_file_chunk_size` - (Optional) For files imported from a data repository, this value determines the stripe count and maximum amount of data per file (in MiB) stored on a single physical disk. Can only be specified with `import_path` argument. Defaults to `1024`. Minimum of `1` and maximum of `512000`.
* `security_group_ids` - (Optional) A list of IDs for the security groups that apply to the specified network interfaces created for file system access. These security groups will apply to all network interfaces.
* `tags` - (Optional) A map of tags to assign to the file system.
* `weekly_maintenance_start_time` - (Optional) The preferred start time (in `d:HH:MM` format) to perform weekly maintenance, in the UTC time zone.
* `deployment_type` - (Optional) - The filesystem deployment type. One of: `SCRATCH_1`, `SCRATCH_2`, `PERSISTENT_1`.
* `kms_key_id` - (Optional) ARN for the KMS Key to encrypt the file system at rest, applicable for `PERSISTENT_1` deployment_type. Defaults to an AWS managed KMS Key.
* `per_unit_storage_throughput` - (Optional) - Describes the amount of read and write throughput for each 1 tebibyte of storage, in MB/s/TiB, required for the `PERSISTENT_1` deployment_type. Valid values for `SSD` storage_type are 50, 100, 200. Valid values for `HDD` storage_type are 12, 40.
* `automatic_backup_retention_days` - (Optional) The number of days to retain automatic backups. Setting this to 0 disables automatic backups. You can retain automatic backups for a maximum of 90 days. only valid for `PERSISTENT_1` deployment_type.
* `storage_type` - (Optional) - The filesystem storage type. Either `SSD` or `HDD`, defaults to `SSD`. `HDD` is only supported on `PERSISTENT_1` deployment types.
* `drive_cache_type` - (Optional) - The type of drive cache used by `PERSISTENT_1` filesystems that are provisioned with `HDD` storage_type. Required for `HDD` storage_type, set to either `READ` or `NONE`.
* `daily_automatic_backup_start_time` - (Optional) A recurring daily time, in the format HH:MM. HH is the zero-padded hour of the day (0-23), and MM is the zero-padded minute of the hour. For example, 05:00 specifies 5 AM daily. only valid for `PERSISTENT_1` deployment_type. Requires `automatic_backup_retention_days` to be set.
* `auto_import_policy` - (Optional) How Amazon FSx keeps your file and directory listings up to date as you add or modify objects in your linked S3 bucket. see [Auto Import Data Repo](https://docs.aws.amazon.com/fsx/latest/LustreGuide/autoimport-data-repo.html) for more details.
* `copy_tags_to_backups` - (Optional) A boolean flag indicating whether tags for the file system should be copied to backups. Applicable for `PERSISTENT_1` deployment_type. The default value is false.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `dns_name` - DNS name for the file system, e.g. `fs-12345678.fsx.us-west-2.amazonaws.com`
* `id` - Identifier of the file system, e.g. `fs-12345678`
* `network_interface_ids` - Set of Elastic Network Interface identifiers from which the file system is accessible. As explained in the [documentation](https://docs.aws.amazon.com/fsx/latest/LustreGuide/mounting-on-premises.html), the first network interface returned is the primary network interface.
* `mount_name` - The value to be used when mounting the filesystem.
* `owner_id` - AWS account identifier that created the file system.
* `vpc_id` - Identifier of the Virtual Private Cloud for the file system.

## Timeouts

`aws_fsx_lustre_file_system` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `30m`) How long to wait for the file system to be created.
* `delete` - (Default `30m`) How long to wait for the file system to be deleted.

## Import

FSx File Systems can be imported using the `id`, e.g.

```
$ terraform import aws_fsx_lustre_file_system.example fs-543ab12b1ca672f33
```

Certain resource arguments, like `security_group_ids`, do not have a FSx API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](/docs/configuration/resources.html#ignore_changes) to hide the difference, e.g.

```hcl
resource "aws_fsx_lustre_file_system" "example" {
  # ... other configuration ...
  security_group_ids = [aws_security_group.example.id]

  # There is no FSx API for reading security_group_ids
  lifecycle {
    ignore_changes = [security_group_ids]
  }
}
```
