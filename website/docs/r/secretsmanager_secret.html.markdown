---
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret"
sidebar_current: "docs-aws-resource-secretsmanager-secret"
description: |-
  Provides a resource to manage AWS Secrets Manager secret metadata
---

# aws_secretsmanager_secret

Provides a resource to manage AWS Secrets Manager secret metadata. To manage a secret value, see the [`aws_secretsmanager_secret_version` resource](/docs/providers/aws/r/secretsmanager_secret_version.html).

## Example Usage

### Basic

```hcl
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}
```

### Rotation Configuration

To enable automatic secret rotation, the Secrets Manager service requires usage of a Lambda function. The [Rotate Secrets section in the Secrets Manager User Guide](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html) provides additional information about deploying a prebuilt Lambda functions for supported credential rotation (e.g. RDS) or deploying a custom Lambda function.

~> **NOTE:** Configuring rotation causes the secret to rotate once as soon as you store the secret. Before you do this, you must ensure that all of your applications that use the credentials stored in the secret are updated to retrieve the secret from AWS Secrets Manager. The old credentials might no longer be usable after the initial rotation and any applications that you fail to update will break as soon as the old credentials are no longer valid.

~> **NOTE:** If you cancel a rotation that is in progress (by removing the `rotation` configuration), it can leave the VersionStage labels in an unexpected state. Depending on what step of the rotation was in progress, you might need to remove the staging label AWSPENDING from the partially created version, specified by the SecretVersionId response value. You should also evaluate the partially rotated new version to see if it should be deleted, which you can do by removing all staging labels from the new version's VersionStage field.

```hcl
resource "aws_secretsmanager_secret" "rotation-example" {
  name                = "rotation-example"
  rotation_lambda_arn = "${aws_lambda_function.example.arn}"

  rotation_rules {
    automatically_after_days = 7
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Specifies the friendly name of the new secret. The secret name can consist of uppercase letters, lowercase letters, digits, and any of the following characters: `/_+=.@-` Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) A description of the secret.
* `kms_key_id` - (Optional) Specifies the ARN or alias of the AWS KMS customer master key (CMK) to be used to encrypt the secret values in the versions stored in this secret. If you don't specify this value, then Secrets Manager defaults to using the AWS account's default CMK (the one named `aws/secretsmanager`). If the default KMS CMK with that name doesn't yet exist, then AWS Secrets Manager creates it for you automatically the first time.
* `policy` - (Optional) A valid JSON document representing a [resource policy](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_resource-based-policies.html). For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](/docs/providers/aws/guides/iam-policy-documents.html).
* `recovery_window_in_days` - (Optional) Specifies the number of days that AWS Secrets Manager waits before it can delete the secret. This value can be `0` to force deletion without recovery or range from `7` to `30` days. The default value is `30`.
* `rotation_lambda_arn` - (Optional) Specifies the ARN of the Lambda function that can rotate the secret.
* `rotation_rules` - (Optional) A structure that defines the rotation configuration for this secret. Defined below.
* `tags` - (Optional) Specifies a key-value map of user-defined tags that are attached to the secret.

### rotation_rules

* `automatically_after_days` - (Required) Specifies the number of days between automatic scheduled rotations of the secret.

## Attribute Reference

* `id` - Amazon Resource Name (ARN) of the secret.
* `arn` - Amazon Resource Name (ARN) of the secret.
* `rotation_enabled` - Specifies whether automatic rotation is enabled for this secret.

## Import

`aws_secretsmanager_secret` can be imported by using the secret Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_secretsmanager_secret.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
