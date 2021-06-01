---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault"
description: |-
  Provides details about an AWS Backup vault.
---

# Data Source: aws_backup_vault

Use this data source to get information on an existing backup vault.

## Example Usage

```hcl
data "aws_backup_vault" "example" {
  name = "example_backup_vault"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the backup vault.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the vault.
* `kms_key_arn` - The server-side encryption key that is used to protect your backups.
* `recovery_points` - The number of recovery points that are stored in a backup vault.
* `tags` - Metadata that you can assign to help organize the resources that you create.
