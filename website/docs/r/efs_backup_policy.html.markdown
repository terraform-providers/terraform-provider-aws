---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_backup_policy"
description: |-
  Provides an Elastic File System (EFS) Backup Policy resource.
---

# Resource: aws_efs_backup_policy

Provides an Elastic File System (EFS) Backup Policy resource.
Backup policies turn automatic backups on or off for an existing file system.

## Example Usage

```terraform
resource "aws_efs_file_system" "fs" {
  creation_token = "my-product"
}

resource "aws_efs_backup_policy" "policy" {
  file_system_id = aws_efs_file_system.fs.id

  backup_policy {
    status = "ENABLED"
  }
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Required) The ID of the EFS file system.
* `backup_policy` - (Required) A backup_policy object (documented below).

### Backup Policy Arguments
For **backup_policy** the following attributes are supported:

* `status` - (Required) A status of the backup policy. Valid values: `ENABLED`, `DISABLED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID that identifies the file system (e.g. fs-ccfc0d65).

## Import

The EFS backup policies can be imported using the `id`, e.g.

```
$ terraform import aws_efs_backup_policy.example fs-6fa144c6
```
