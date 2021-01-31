---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecr_authorization_token"
description: |-
    Provides details about an ECR Authorization Token
---

# Data Source: aws_ecr_authorization_token

The ECR Authorization Token data source allows the authorization token, proxy endpoint, token expiration date, user name and password to be retrieved for an ECR repository.

## Example Usage

```hcl
data "aws_ecr_authorization_token" "token" {
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) AWS account ID of the ECR Repository. If not specified the default account is assumed.

## Attributes Reference

In addition to the argument above, the following attributes are exported:

* `authorization_token` - Temporary IAM authentication credentials to access the ECR repository encoded in base64 in the form of `user_name:password`.
* `expires_at` - The time in UTC RFC3339 format when the authorization token expires.
* `id` - Region of the authorization token.
* `password` - Password decoded from the authorization token.
* `proxy_endpoint` - The registry URL to use in the docker login command.
* `user_name` - User name decoded from the authorization token.
