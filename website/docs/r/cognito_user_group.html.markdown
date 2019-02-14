---
layout: "aws"
page_title: "AWS: aws_cognito_user_group"
sidebar_current: "docs-aws-resource-cognito-user-group"
description: |-
  Provides a Cognito User Group resource.
---

# aws_cognito_user_group

Provides a Cognito User Group resource.

## Example Usage

```hcl
resource "aws_cognito_user_pool" "main" {
  name = "identity pool"
}

resource "aws_iam_role" "group_role" {
  name = "user-group-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "cognito-identity.amazonaws.com:aud": "us-east-1:12345678-dead-beef-cafe-123456790ab"
        },
        "ForAnyValue:StringLike": {
          "cognito-identity.amazonaws.com:amr": "authenticated"
        }
      }
    }
  ]
}
EOF
}

resource "aws_cognito_user_group" "main" {
  name         = "user-group"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
  description  = "Managed by Terraform"
  precedence   = 42
  role_arn     = "${aws_iam_role.group_role.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the user group.
* `user_pool_id` - (Required) The user pool ID.
* `description` - (Optional) The description of the user group.
* `precedence` - (Optional) The precedence of the user group.
* `role_arn` - (Optional) The ARN of the IAM role to be associated with the user group.

## Import

Cognito User Groups can be imported using the `user_pool_id`/`name` attributes concatenated, e.g.

```
$ terraform import aws_cognito_user_group.group us-east-1_vG78M4goG/user-group
```
