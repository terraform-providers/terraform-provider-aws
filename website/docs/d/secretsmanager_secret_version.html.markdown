---
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_version"
sidebar_current: "docs-aws-datasource-secretsmanager-secret-version"
description: |-
  Retrieve information about a Secrets Manager secret version including its secret value
---

# Data Source: aws_secretsmanager_secret_version

Retrieve information about a Secrets Manager secret version, including its secret value. To retrieve secret metadata, see the [`aws_secretsmanager_secret` data source](/docs/providers/aws/d/secretsmanager_secret.html).

## Example Usage

### Retrieve Current Secret Version

By default, this data sources retrieves information based on the `AWSCURRENT` staging label.

```hcl
data "aws_secretsmanager_secret_version" "example" {
  secret_id = "${data.aws_secretsmanager_secret.example.id}"
}
```

### Retrieve Specific Secret Version

```hcl
data "aws_secretsmanager_secret_version" "by-version-stage" {
  secret_id     = "${data.aws_secretsmanager_secret.example.id}"
  version_stage = "example"
}
```

## Argument Reference

* `secret_id` - (Required) Specifies the secret containing the version that you want to retrieve. You can specify either the Amazon Resource Name (ARN) or the friendly name of the secret.
* `version_id` - (Optional) Specifies the unique identifier of the version of the secret that you want to retrieve. Overrides `version_stage`.
* `version_stage` - (Optional) Specifies the secret version that you want to retrieve by the staging label attached to the version. Defaults to `AWSCURRENT`.

## Attributes Reference

* `arn` - The ARN of the secret.
* `id` - The unique identifier of this version of the secret.
* `secret_string` - The decrypted part of the protected secret information that was originally provided as a string.
* `secret_binary` - The decrypted part of the protected secret information that was originally provided as a binary. Base64 encoded.
* `version_id` - The unique identifier of this version of the secret.
