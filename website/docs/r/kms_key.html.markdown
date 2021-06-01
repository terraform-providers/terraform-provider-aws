---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_key"
description: |-
  Provides a KMS customer master key.
---

# Resource: aws_kms_key

Provides a KMS customer master key.

## Example Usage

```hcl
resource "aws_kms_key" "a" {
  description             = "KMS key 1"
  deletion_window_in_days = 10
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the key as viewed in AWS console.
* `key_usage` - (Optional) Specifies the intended use of the key. Valid values: `ENCRYPT_DECRYPT` or `SIGN_VERIFY`.
Defaults to `ENCRYPT_DECRYPT`.
* `customer_master_key_spec` - (Optional) Specifies whether the key contains a symmetric key or an asymmetric key pair and the encryption algorithms or signing algorithms that the key supports.
Valid values: `SYMMETRIC_DEFAULT`,  `RSA_2048`, `RSA_3072`, `RSA_4096`, `ECC_NIST_P256`, `ECC_NIST_P384`, `ECC_NIST_P521`, or `ECC_SECG_P256K1`. Defaults to `SYMMETRIC_DEFAULT`. For help with choosing a key spec, see the [AWS KMS Developer Guide](https://docs.aws.amazon.com/kms/latest/developerguide/symm-asymm-choose.html).
* `policy` - (Optional) A valid policy JSON document. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `deletion_window_in_days` - (Optional) Duration in days after which the key is deleted after destruction of the resource, must be between 7 and 30 days. Defaults to 30 days.
* `is_enabled` - (Optional) Specifies whether the key is enabled. Defaults to true.
* `enable_key_rotation` - (Optional) Specifies whether [key rotation](http://docs.aws.amazon.com/kms/latest/developerguide/rotate-keys.html) is enabled. Defaults to false.
* `tags` - (Optional) A map of tags to assign to the object.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the key.
* `key_id` - The globally unique identifier for the key.

## Import

KMS Keys can be imported using the `id`, e.g.

```
$ terraform import aws_kms_key.a 1234abcd-12ab-34cd-56ef-1234567890ab
```
