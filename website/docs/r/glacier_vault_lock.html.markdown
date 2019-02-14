---
layout: "aws"
page_title: "AWS: aws_glacier_vault_lock"
sidebar_current: "docs-aws-resource-glacier-vault-lock"
description: |-
  Manages a Glacier Vault Lock.
---

# aws_glacier_vault_lock

Manages a Glacier Vault Lock. You can refer to the [Glacier Developer Guide](https://docs.aws.amazon.com/amazonglacier/latest/dev/vault-lock.html) for a full explanation of the Glacier Vault Lock functionality.

~> **NOTE:** This resource allows you to test Glacier Vault Lock policies by setting the `complete_lock` argument to `false`. When testing policies in this manner, the Glacier Vault Lock automatically expires after 24 hours and Terraform will show this resource as needing recreation after that time. To permanently apply the policy, set the `complete_lock` argument to `true`. When changing `complete_lock` to `true`, it is expected the resource will show as recreating.

!> **WARNING:** Once a Glacier Vault Lock is completed, it is immutable. The deletion of the Glacier Vault Lock is not be possible and attempting to remove it from Terraform will return an error. Set the `ignore_deletion_error` argument to `true` and apply this configuration before attempting to delete this resource via Terraform or use `terraform state rm` to remove this resource from Terraform management.

## Example Usage

### Testing Glacier Vault Lock Policy

```hcl
resource "aws_glacier_vault" "example" {
  name = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    actions   = ["glacier:DeleteArchive"]
    effect    = "Deny"
    resources = ["${aws_glacier_vault.example.arn}"]

    condition {
      test     = "NumericLessThanEquals"
      variable = "glacier:ArchiveAgeinDays"
      values   = ["365"]
    }
  }
}

resource "aws_glacier_vault_lock" "example" {
  complete_lock = false
  policy        = "${data.aws_iam_policy_document.example.json}"
  vault_name    = "${aws_glacier_vault.example.name}"
}
```

### Permanently Applying Glacier Vault Lock Policy

```hcl
resource "aws_glacier_vault_lock" "example" {
  complete_lock = true
  policy        = "${data.aws_iam_policy_document.example.json}"
  vault_name    = "${aws_glacier_vault.example.name}"
}
```

## Argument Reference

The following arguments are supported:

* `complete_lock` - (Required) Boolean whether to permanently apply this Glacier Lock Policy. Once completed, this cannot be undone. If set to `false`, the Glacier Lock Policy remains in a testing mode for 24 hours. After that time, the Glacier Lock Policy is automatically removed by Glacier and the Terraform resource will show as needing recreation. Changing this from `false` to `true` will show as resource recreation, which is expected. Changing this from `true` to `false` is not possible unless the Glacier Vault is recreated at the same time.
* `policy` - (Required) JSON string containing the IAM policy to apply as the Glacier Vault Lock policy.
* `vault_name` - (Required) The name of the Glacier Vault.
* `ignore_deletion_error` - (Optional) Allow Terraform to ignore the error returned when attempting to delete the Glacier Lock Policy. This can be used to delete or recreate the Glacier Vault via Terraform, for example, if the Glacier Vault Lock policy permits that action. This should only be used in conjunction with `complete_lock` being set to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Glacier Vault name.

## Import

Glacier Vault Locks can be imported using the Glacier Vault name, e.g.

```
$ terraform import aws_glacier_vault_lock.example example-vault
```
