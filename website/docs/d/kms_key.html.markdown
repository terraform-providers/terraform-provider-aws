---
layout: "aws"
page_title: "AWS: aws_kms_key"
sidebar_current: "docs-aws-datasource-kms-key"
description: |-
  Get information on a AWS Key Management Service (KMS) Key
---

# aws_kms_key

Use this data source to get detailed information about 
the specified KMS Key with flexible key id input. 
This can be useful to reference key alias 
without having to hard code the ARN as input.

## Example Usage

```hcl
data "aws_kms_key" "foo" {
  key_id = "alias/my-key"
}

data "aws_kms_key" "foo" {
  key_id = "1234abcd-12ab-34cd-56ef-1234567890ab"
}

data "aws_kms_key" "foo" {
  key_id = "arn:aws:kms:us-east-1:111122223333:alias/my-key"
}

data "aws_kms_key" "foo" {
  key_id = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
}
```

## Argument Reference

* `key_id` - (Required) Key identifier which can be one of the following format:
  * Key ID. E.g: `1234abcd-12ab-34cd-56ef-1234567890ab`
  * Key ARN. E.g.: `arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab`
  * Alias name. E.g.: `alias/my-key`
  * Alias ARN: E.g.: `arn:aws:kms:us-east-1:111122223333:alias/my-key`
* `grant_tokens` - (Optional) List of grant tokens

## Attributes Reference

* `id`: The globally unique identifier for the key
* `arn`: The Amazon Resource Name (ARN) of the key
* `aws_account_id`: The twelve-digit account ID of the AWS account that owns the key
* `creation_date`: The date and time when the key was created
* `deletion_date`: The date and time after which AWS KMS deletes the key. This value is present only when `key_state` is `PendingDeletion`, otherwise this value is 0
* `description`: The description of the key.
* `enabled`: Specifies whether the key is enabled. When `key_state` is `Enabled` this value is true, otherwise it is false
* `expiration_model`: Specifies whether the Key's key material expires. This value is present only when `origin` is `EXTERNAL`, otherwise this value is empty
* `key_manager`: The key's manager
* `key_state`: The state of the key
* `key_usage`: Currently the only allowed value is `ENCRYPT_DECRYPT`
* `origin`: When this value is `AWS_KMS`, AWS KMS created the key material. When this value is `EXTERNAL`, the key material was imported from your existing key management infrastructure or the CMK lacks key material
* `valid_to`: The time at which the imported key material expires. This value is present only when `origin` is `EXTERNAL` and whose `expiration_model` is `KEY_MATERIAL_EXPIRES`, otherwise this value is 0
