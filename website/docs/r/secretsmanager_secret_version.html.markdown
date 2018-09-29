---
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_version"
sidebar_current: "docs-aws-resource-secretsmanager-secret-version"
description: |-
  Provides a resource to manage AWS Secrets Manager secret version including its secret value
---

# aws_secretsmanager_secret_version

Provides a resource to manage AWS Secrets Manager secret version including its secret value. To manage secret metadata, see the [`aws_secretsmanager_secret` resource](/docs/providers/aws/r/secretsmanager_secret.html).

~> **NOTE:** If the `AWSCURRENT` staging label is present on this version during resource deletion, that label cannot be removed and will be skipped to prevent errors when fully deleting the secret. That label will leave this secret version active even after the resource is deleted from Terraform unless the secret itself is deleted. Move the `AWSCURRENT` staging label before or after deleting this resource from Terraform to fully trigger version deprecation if necessary.

## Example Usage

### Simple String Value

```hcl
resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = "${aws_secretsmanager_secret.example.id}"
  secret_string = "example-string-to-protect"
}
```

### Key-Value Pairs

Secrets Manager also accepts key-value pairs in JSON.

```hcl
# The map here can come from other supported configurations
# like locals, resource attribute, map() built-in, etc.
variable "example" {
  default = {
    key1 = "value1"
    key2 = "value2"
  }

  type = "map"
}

resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = "${aws_secretsmanager_secret.example.id}"
  secret_string = "${jsonencode(var.example)}"
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) Specifies the secret to which you want to add a new version. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret. The secret must already exist.
* `secret_string` - (Optional) Specifies text data that you want to encrypt and store in this version of the secret.
* `generate_random_password` - (Optional) You can use this to generate random text data instead of specifying the text data in `secret_string`. Defined below
~> **NOTE:** You should use either one between `secret_string` and `generate_random_password`.
* `version_stages` - (Optional) Specifies a list of staging labels that are attached to this version of the secret. A staging label must be unique to a single version of the secret. If you specify a staging label that's already associated with a different version of the same secret then that staging label is automatically removed from the other version and attached to this version. If you do not specify a value, then AWS Secrets Manager automatically moves the staging label `AWSCURRENT` to this new version on creation.

~> **NOTE:** If `version_stages` is configured, you must include the `AWSCURRENT` staging label if this secret version is the only version or if the label is currently present on this secret version, otherwise Terraform will show a perpetual difference.


## generate_random_password

Full details on the core parameters and impacts are in the API Docs: [GetRandomPassword](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetRandomPassword.html)

```hcl
resource "aws_secretsmanager_secret_version" "example" {
  secret_id = "${aws_secretsmanager_secret.example.id}"
  generate_random_password {
    exclude_characters = false
    exclude_lowercase = false
    exclude_numbers = false
    exclude_punctuation = false
    exclude_uppercase = false
    include_space = false
    password_length = 32
    require_each_included_type = true
  }
}
```

* `exclude_characters` - (Optional) A string that includes characters that should not be included in the generated password. The default is that all characters from the included sets can be used.
* `exclude_lowercase` - (Optional) Specifies that the generated password should not include lowercase letters. The default if you do not include this switch parameter is that lowercase letters can be included.
* `exclude_numbers` - (Optional) Specifies that the generated password should not include digits. The default if you do not include this switch parameter is that digits can be included.
* `exclude_punctuation` - (Optional) Specifies that the generated password should not include punctuation characters. The default if you do not include this switch parameter is that punctuation characters can be included.
* `exclude_uppercase` - (Optional) Specifies that the generated password should not include uppercase letters. The default if you do not include this switch parameter is that uppercase letters can be included.
* `include_space` - (Optional) Specifies that the generated password can include the space character. The default if you do not include this switch parameter is that the space character is not included.
* `password_length` - (Optional) The desired length of the generated password. The default value if you do not include this parameter is 32 characters.
* `require_each_included_type` - (Optional) A boolean value that specifies whether the generated password must include at least one of every allowed character type. The default value is True and the operation requires at least one of every character type.


## Attribute Reference

* `arn` - The ARN of the secret.
* `id` - A pipe delimited combination of secret ID and version ID.
* `version_id` - The unique identifier of the version of the secret.

## Import

`aws_secretsmanager_secret_version` can be imported by using the secret ID and version ID, e.g.

```
$ terraform import aws_secretsmanager_secret.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456|xxxxx-xxxxxxx-xxxxxxx-xxxxx
```
