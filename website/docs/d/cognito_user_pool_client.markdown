---
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client"
sidebar_current: "docs-aws-cognito-user-pool-client"
description: |-
  Get information on a Cognito User Pool Client.
---

# Data Source: aws_cognito_user_pool_client

Use this data source to get information on a Cognito User Pool Client.

## Example Usage

```hcl

data "aws_cognito_user_pool_client" "selected" {
  name         = "${var.cognito_user_pool_client_name}"
  user_pool_id = "${var.cognito_user_pool_id}"
}

```

## Argument Reference

- `name` - (required) Name of the cognito user pool client.
- `user_pool_id` - (required) Id of the cognito user pool.

## Attributes Reference

- `client_id` - Client Id of the user pool client.
- `client_secret` - Client Secret of the user pool client.
