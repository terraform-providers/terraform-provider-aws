---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_vault_policy"
description: |-
  Provides an AWS Backup vault policy resource.
---

# Resource: aws_backup_vault_policy

Provides an AWS Backup vault policy resource.

## Example Usage

```hcl
resource "aws_backup_vault" "example" {
  name = "example"
}

resource "aws_backup_vault_policy" "example" {
  backup_vault_name = aws_backup_vault.example.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
		"backup:DescribeBackupVault",
		"backup:DeleteBackupVault",
		"backup:PutBackupVaultAccessPolicy",
		"backup:DeleteBackupVaultAccessPolicy",
		"backup:GetBackupVaultAccessPolicy",
		"backup:StartBackupJob",
		"backup:GetBackupVaultNotifications",
		"backup:PutBackupVaultNotifications"
      ],
      "Resource": "${aws_backup_vault.example.arn}"
    }
  ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `backup_vault_name` - (Required) Name of the backup vault to add policy for.
* `policy` - (Required) The backup vault access policy document in JSON format.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the vault.
* `backup_vault_arn` - The ARN of the vault.

## Import

Backup vault policy can be imported using the `name`, e.g.

```
$ terraform import aws_backup_vault_policy.test TestVault
```
