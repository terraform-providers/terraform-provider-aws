---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_public_key"
description: |-
  Get information on a KMS public key
---

# aws_kms_public_key

Use this data source to get the public key about the specified KMS Key with flexible key id input. This can be useful to reference key alias without having to hard code the ARN as input.

## Example Usage

```terraform
data "aws_kms_public_key" "by_alias" {
  key_id = "alias/my-key"
}

data "aws_kms_public_key" "by_id" {
  key_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}

data "aws_kms_public_key" "by_alias_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:alias/my-key"
}

data "aws_kms_public_key" "by_key_arn" {
  key_id = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

## Argument Reference

The following arguments are supported:

* `key_id` - (Required) Key identifier which can be one of the following format:
    * Key ID. E.g - `1234abcd-12ab-34cd-56ef-1234567890ab`
    * Key ARN. E.g. - `arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab`
    * Alias name. E.g. - `alias/my-key`
    * Alias ARN - E.g. - `arn:aws:kms:us-east-1:111122223333:alias/my-key`
* `grant_tokens` - (Optional) List of grant tokens

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Key ARN of the asymmetric CMK from which the public key was downloaded.
* `customer_master_key_spec` - Type of the public key that was downloaded.
* `encryption_algorithms` - Encryption algorithms that AWS KMS supports for this key. Only set when the `key_usage` of the public key is `ENCRYPT_DECRYPT`.
* `id` - Key ARN of the asymmetric CMK from which the public key was downloaded.
* `key_usage` - Permitted use of the public key. Valid values are `ENCRYPT_DECRYPT` or `SIGN_VERIFY`
* `public_key` - Exported public key. The value is a DER-encoded X.509 public key, also known as SubjectPublicKeyInfo (SPKI), as defined in [RFC 5280](https://tools.ietf.org/html/rfc5280). The value is Base64-encoded.
* `signing_algorithms` - Signing algorithms that AWS KMS supports for this key. Only set when the `key_usage` of the public key is `SIGN_VERIFY`.
