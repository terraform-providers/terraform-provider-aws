---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_rotation"
description: |-
  Provides a resource to manage AWS Secrets Manager secret rotation
---

# Resource: aws_secretsmanager_secret_rotation

Provides a resource to manage AWS Secrets Manager secret rotation. To manage a secret, see the [`aws_secretsmanager_secret` resource](/docs/providers/aws/r/secretsmanager_secret.html). To manage a secret value, see the [`aws_secretsmanager_secret_version` resource](/docs/providers/aws/r/secretsmanager_secret_version.html).

## Example Usage

### Basic

```hcl
resource "aws_secretsmanager_secret_rotation" "example" {
  secret_id           = aws_secretsmanager_secret.example.id
  rotation_lambda_arn = aws_lambda_function.example.arn

  rotation_rules {
    automatically_after_days = 30
  }
}
```

### Rotation Configuration

To enable automatic secret rotation, the Secrets Manager service requires usage of a Lambda function. The [Rotate Secrets section in the Secrets Manager User Guide](https://docs.aws.amazon.com/secretsmanager/latest/userguide/rotating-secrets.html) provides additional information about deploying a prebuilt Lambda functions for supported credential rotation (e.g. RDS) or deploying a custom Lambda function.

~> **NOTE:** Configuring rotation causes the secret to rotate once as soon as you enable rotation. Before you do this, you must ensure that all of your applications that use the credentials stored in the secret are updated to retrieve the secret from AWS Secrets Manager. The old credentials might no longer be usable after the initial rotation and any applications that you fail to update will break as soon as the old credentials are no longer valid.

~> **NOTE:** If you cancel a rotation that is in progress (by removing the `rotation` configuration), it can leave the VersionStage labels in an unexpected state. Depending on what step of the rotation was in progress, you might need to remove the staging label AWSPENDING from the partially created version, specified by the SecretVersionId response value. You should also evaluate the partially rotated new version to see if it should be deleted, which you can do by removing all staging labels from the new version's VersionStage field.

## Argument Reference

The following arguments are supported:

* `secret_id` - (Required) Specifies the secret to which you want to add a new version. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret. The secret must already exist.
* `rotation_lambda_arn` - (Required) Specifies the ARN of the Lambda function that can rotate the secret.
* `rotation_rules` - (Required) A structure that defines the rotation configuration for this secret. Defined below.

### rotation_rules

* `automatically_after_days` - (Required) Specifies the number of days between automatic scheduled rotations of the secret.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the secret.
* `arn` - Amazon Resource Name (ARN) of the secret.
* `rotation_enabled` - Specifies whether automatic rotation is enabled for this secret.

## Import

`aws_secretsmanager_secret_rotation` can be imported by using the secret Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_secretsmanager_secret_rotation.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
