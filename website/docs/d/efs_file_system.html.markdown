---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_file_system"
description: |-
  Provides an Elastic File System (EFS) File System data source.
---

# Data Source: aws_efs_file_system

Provides information about an Elastic File System (EFS) File System.

## Example Usage

```hcl
variable "file_system_id" {
  type    = string
  default = ""
}

data "aws_efs_file_system" "by_id" {
  file_system_id = var.file_system_id
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Optional) The ID that identifies the file system (e.g. fs-ccfc0d65).
* `creation_token` - (Optional) Restricts the list to the file system with this creation token.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `dns_name` - The DNS name for the filesystem per [documented convention](http://docs.aws.amazon.com/efs/latest/ug/mounting-fs-mount-cmd-dns-name.html).
* `encrypted` - Whether EFS is encrypted.
* `kms_key_id` - The ARN for the KMS encryption key.
* `lifecycle_policy` - A file system [lifecycle policy](https://docs.aws.amazon.com/efs/latest/ug/API_LifecyclePolicy.html) object.
* `performance_mode` - The file system performance mode.
* `provisioned_throughput_in_mibps` - The throughput, measured in MiB/s, that you want to provision for the file system.
* `tags` -A map of tags to assign to the file system.
* `throughput_mode` - Throughput mode for the file system.
* `size_in_bytes` - The current byte count used by the file system.
