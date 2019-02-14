---
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret"
sidebar_current: "docs-aws-datasource-secretsmanager-secret"
description: |-
  Retrieve metadata information about a Secrets Manager secret
---

# Data Source: aws_secretsmanager_secret

Retrieve metadata information about a Secrets Manager secret. To retrieve a secret value, see the [`aws_secretsmanager_secret_version` data source](/docs/providers/aws/d/secretsmanager_secret_version.html).

## Example Usage

### ARN

```hcl
data "aws_secretsmanager_secret" "by-arn" {
  arn = "arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456"
}
```

### Name

```hcl
data "aws_secretsmanager_secret" "by-name" {
  name = "example"
}
```

## Argument Reference

* `arn` - (Optional) The Amazon Resource Name (ARN) of the secret to retrieve.
* `name` - (Optional) The name of the secret to retrieve.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the secret.
* `description` - A description of the secret.
* `kms_key_id` - The Key Management Service (KMS) Customer Master Key (CMK) associated with the secret.
* `id` - The Amazon Resource Name (ARN) of the secret.
* `rotation_enabled` - Whether rotation is enabled or not.
* `rotation_lambda_arn` - Rotation Lambda function Amazon Resource Name (ARN) if rotation is enabled.
* `rotation_rules` - Rotation rules if rotation is enabled.
* `tags` - Tags of the secret.
* `policy` - The resource-based policy document that's attached to the secret.
