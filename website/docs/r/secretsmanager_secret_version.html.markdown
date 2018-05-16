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

```hcl
resource "aws_secretsmanager_secret_version" "example" {
  secret_id     = "${aws_secretsmanager_secret.example.id}"
  secret_string = "example-string-to-protect"
}
```

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) Specifies the secret to which you want to add a new version. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret. The secret must already exist.
* `secret_string` - (Required) Specifies text data that you want to encrypt and store in this version of the secret.
* `version_stages` - (Optional) Specifies a list of staging labels that are attached to this version of the secret. A staging label must be unique to a single version of the secret. If you specify a staging label that's already associated with a different version of the same secret then that staging label is automatically removed from the other version and attached to this version. If you do not specify a value, then AWS Secrets Manager automatically moves the staging label `AWSCURRENT` to this new version on creation.

~> **NOTE:** If `version_stages` is configured, you must include the `AWSCURRENT` staging label if this secret version is the only version or if the label is currently present on this secret version, otherwise Terraform will show a perpetual difference.

## Attribute Reference

* `id` - A pipe delimited combination of secret ID and version ID
* `version_id` - The unique identifier of the version of the secret.

## Import

`aws_secretsmanager_secret_version` can be imported by using the secret ID and version ID, e.g.

```
$ terraform import aws_secretsmanager_secret.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456|xxxxx-xxxxxxx-xxxxxxx-xxxxx
```
