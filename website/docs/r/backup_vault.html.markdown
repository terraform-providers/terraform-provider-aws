---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault"
description: |-
  Provides an AWS Backup vault resource.
---

# Resource: aws_backup_vault

Provides an AWS Backup vault resource.

## Example Usage

```hcl
resource "aws_backup_vault" "example" {
  name        = "example_backup_vault"
  kms_key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the backup vault to create.
* `tags` - (Optional) Metadata that you can assign to help organize the resources that you create.
* `kms_key_arn` - (Optional) The server-side encryption key that is used to protect your backups.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the vault.
* `arn` - The ARN of the vault.
* `recovery_points` - The number of recovery points that are stored in a backup vault.

## Import

Backup vault can be imported using the `name`, e.g.

```
$ terraform import aws_backup_vault.test-vault TestVault
```
