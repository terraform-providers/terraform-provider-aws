---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_rotation"
description: |-
  Retrieve information about a Secrets Manager secret rotation configuration
---

# Data Source: aws_secretsmanager_secret_rotation

Retrieve information about a Secrets Manager secret rotation. To retrieve secret metadata, see the [`aws_secretsmanager_secret` data source](/docs/providers/aws/d/secretsmanager_secret.html). To retrieve a secret value, see the [`aws_secretsmanager_secret_version` data source](/docs/providers/aws/d/secretsmanager_secret_version.html).

## Example Usage

### Retrieve Secret Rotation Configuration

```hcl
data "aws_secretsmanager_secret_rotation" "example" {
  secret_id = data.aws_secretsmanager_secret.example.id
}
```

## Argument Reference

* `secret_id` - (Required) Specifies the secret containing the version that you want to retrieve. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret.

## Attributes Reference

* `rotation_enabled` - The ARN of the secret.
* `rotation_lambda_arn` - The decrypted part of the protected secret information that was originally provided as a string.
* `rotation_rules` - The decrypted part of the protected secret information that was originally provided as a binary. Base64 encoded.
