---
subcategory: "Cognito"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client"
description: |-
  Provides a Cognito User Pool Client resource.
---

# Resource: aws_cognito_user_pool_client

Provides a Cognito User Pool Client resource.

## Example Usage

### Create a basic user pool client

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}

resource "aws_cognito_user_pool_client" "client" {
  name = "client"

  user_pool_id = aws_cognito_user_pool.pool.id
}
```

### Create a user pool client with no SRP authentication

```hcl
resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}

resource "aws_cognito_user_pool_client" "client" {
  name = "client"

  user_pool_id = aws_cognito_user_pool.pool.id

  generate_secret     = true
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}
```

### Create a user pool client with pinpoint analytics

```hcl
data "aws_caller_identity" "current" {}

resource "aws_cognito_user_pool" "test" {
  name = "pool"
}

resource "aws_pinpoint_app" "test" {
  name = "pinpoint"
}

resource "aws_iam_role" "test" {
  name = "role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "cognito-idp.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "role_policy"
  role = aws_iam_role.test.id

  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "mobiletargeting:UpdateEndpoint",
        "mobiletargeting:PutItems"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:mobiletargeting:*:${data.aws_caller_identity.current.account_id}:apps/${aws_pinpoint_app.test.application_id}*"
    }
  ]
}
EOF
}

resource "aws_cognito_user_pool_client" "test" {
  name         = "pool_client"
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id   = aws_pinpoint_app.test.application_id
    external_id      = "some_id"
    role_arn         = aws_iam_role.test.arn
    user_data_shared = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `allowed_oauth_flows` - (Optional) List of allowed OAuth flows (code, implicit, client_credentials).
* `allowed_oauth_flows_user_pool_client` - (Optional) Whether the client is allowed to follow the OAuth protocol when interacting with Cognito user pools.
* `allowed_oauth_scopes` - (Optional) List of allowed OAuth scopes (phone, email, openid, profile, and aws.cognito.signin.user.admin).
* `callback_urls` - (Optional) List of allowed callback URLs for the identity providers.
* `default_redirect_uri` - (Optional) The default redirect URI. Must be in the list of callback URLs.
* `explicit_auth_flows` - (Optional) List of authentication flows (ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY,  USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_USER_SRP_AUTH, ALLOW_REFRESH_TOKEN_AUTH).
* `generate_secret` - (Optional) Should an application secret be generated.
* `logout_urls` - (Optional) List of allowed logout URLs for the identity providers.
* `name` - (Required) The name of the application client.
* `prevent_user_existence_errors` - (Optional) Choose which errors and responses are returned by Cognito APIs during authentication, account confirmation, and password recovery when the user does not exist in the user pool. When set to `ENABLED` and the user does not exist, authentication returns an error indicating either the username or password was incorrect, and account confirmation and password recovery return a response indicating a code was sent to a simulated destination. When set to `LEGACY`, those APIs will return a `UserNotFoundException` exception if the user does not exist in the user pool.
* `read_attributes` - (Optional) List of user pool attributes the application client can read from.
* `refresh_token_validity` - (Optional) The time limit in days refresh tokens are valid for.
* `supported_identity_providers` - (Optional) List of provider names for the identity providers that are supported on this client.
* `user_pool_id` - (Required) The user pool the client belongs to.
* `write_attributes` - (Optional) List of user pool attributes the application client can write to.
* `analytics_configuration` - (Optional) The Amazon Pinpoint analytics configuration for collecting metrics for this user pool.

### Analytics Configuration

* `application_id` - (Required) The application ID for an Amazon Pinpoint application.
* `external_id`  - (Required) An ID for the Analytics Configuration.
* `role_arn` - (Required) The ARN of an IAM role that authorizes Amazon Cognito to publish events to Amazon Pinpoint analytics.
* `user_data_shared` (Optional) If set to `true`, Amazon Cognito will include user data in the events it publishes to Amazon Pinpoint analytics.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id of the user pool client.
* `client_secret` - The client secret of the user pool client.

## Import

Cognito User Pool Clients can be imported using the `id` of the Cognito User Pool, and the `id` of the Cognito User Pool Client, e.g.

```
$ terraform import aws_cognito_user_pool_client.client <user_pool_id>/<user_pool_client_id>
```
