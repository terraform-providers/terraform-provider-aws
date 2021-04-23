---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_external_key"
description: |-
  Manages a KMS Customer Master Key that uses external key material
---

# Resource: aws_kms_external_key

Manages a KMS Customer Master Key that uses external key material. To instead manage a KMS Customer Master Key where AWS automatically generates and potentially rotates key material, see the [`aws_kms_key` resource](/docs/providers/aws/r/kms_key.html).

~> **Note:** All arguments including the key material will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

## Example Usage

```hcl
resource "aws_kms_external_key" "example" {
  description = "KMS EXTERNAL for AMI encryption"
}
```

## Argument Reference

The following arguments are supported:

* `deletion_window_in_days` - (Optional) Duration in days after which the key is deleted after destruction of the resource. Must be between `7` and `30` days. Defaults to `30`.
* `description` - (Optional) Description of the key.
* `enabled` - (Optional) Specifies whether the key is enabled. Keys pending import can only be `false`. Imported keys default to `true` unless expired.
* `key_material_base64` - (Optional) Base64 encoded 256-bit symmetric encryption key material to import. The CMK is permanently associated with this key material. The same key material can be reimported, but you cannot import different key material.
* `policy` - (Optional) A key policy JSON document. If you do not provide a key policy, AWS KMS attaches a default key policy to the CMK.
* `tags` - (Optional) A key-value map of tags to assign to the key.
* `valid_to` - (Optional) Time at which the imported key material expires. When the key material expires, AWS KMS deletes the key material and the CMK becomes unusable. If not specified, key material does not expire. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the key.
* `expiration_model` - Whether the key material expires. Empty when pending key material import, otherwise `KEY_MATERIAL_EXPIRES` or `KEY_MATERIAL_DOES_NOT_EXPIRE`.
* `id` - The unique identifier for the key.
* `key_state` - The state of the CMK.
* `key_usage` - The cryptographic operations for which you can use the CMK.

## Import

KMS External Keys can be imported using the `id`, e.g.

```
$ terraform import aws_kms_external_key.a arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
```
