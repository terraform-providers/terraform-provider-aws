---
subcategory: "Secrets Manager"
layout: "aws"
page_title: "AWS: aws_secretsmanager_secret_policy"
description: |-
  Provides a resource to manage AWS Secrets Manager secret policy
---

# Resource: aws_secretsmanager_secret_policy

Provides a resource to manage AWS Secrets Manager secret policy.

## Example Usage

### Basic

```hcl
resource "aws_secretsmanager_secret" "example" {
  name = "example"
}

resource "aws_secretsmanager_secret_policy" "example" {
  secret_arn = aws_secretsmanager_secret.example.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "*"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `secret_arn` - (Required) Secret ARN.
* `policy` - (Required) A valid JSON document representing a [resource policy](https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_resource-based-policies.html). For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `block_public_policy` - (Optional) Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.

## Attribute Reference

* `id` - Amazon Resource Name (ARN) of the secret.

## Import

`aws_secretsmanager_secret_policy` can be imported by using the secret Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_secretsmanager_secret_policy.example arn:aws:secretsmanager:us-east-1:123456789012:secret:example-123456
```
