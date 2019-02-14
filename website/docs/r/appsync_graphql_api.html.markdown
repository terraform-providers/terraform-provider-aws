---
layout: "aws"
page_title: "AWS: aws_appsync_graphql_api"
sidebar_current: "docs-aws-resource-appsync-graphql-api"
description: |-
  Provides an AppSync GraphQL API.
---

# aws_appsync_graphql_api

Provides an AppSync GraphQL API.

## Example Usage

### API Key Authentication

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}
```

### AWS Cognito User Pool Authentication

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AMAZON_COGNITO_USER_POOLS"
  name                = "example"

  user_pool_config {
    aws_region     = "${data.aws_region.current.name}"
    default_action = "DENY"
    user_pool_id   = "${aws_cognito_user_pool.example.id}"
  }
}
```

### AWS IAM Authentication

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "AWS_IAM"
  name                = "example"
}
```

### OpenID Connect Authentication

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "OPENID_CONNECT"
  name                = "example"

  openid_connect_config {
    issuer = "https://example.com"
  }
}
```

### Enabling Logging

```hcl
resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
        "Effect": "Allow",
        "Principal": {
            "Service": "appsync.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAppSyncPushToCloudWatchLogs"
  role       = "${aws_iam_role.example.name}"
}

resource "aws_appsync_graphql_api" "example" {
  # ... other configuration ...

  log_config {
    cloudwatch_logs_role_arn = "${aws_iam_role.example.arn}"
    field_log_level          = "ERROR"
  }
}
```

## Argument Reference

The following arguments are supported:

* `authentication_type` - (Required) The authentication type. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`
* `name` - (Required) A user-supplied name for the GraphqlApi.
* `log_config` - (Optional) Nested argument containing logging configuration. Defined below.
* `openid_connect_config` - (Optional) Nested argument containing OpenID Connect configuration. Defined below.
* `user_pool_config` - (Optional) The Amazon Cognito User Pool configuration. Defined below.

### log_config

The following arguments are supported:

* `cloudwatch_logs_role_arn` - (Required) Amazon Resource Name of the service role that AWS AppSync will assume to publish to Amazon CloudWatch logs in your account.
* `field_log_level` - (Required) Field logging level. Valid values: `ALL`, `ERROR`, `NONE`.

### openid_connect_config

The following arguments are supported:

* `issuer` - (Required) Issuer for the OpenID Connect configuration. The issuer returned by discovery MUST exactly match the value of iss in the ID Token.
* `auth_ttl` - (Optional) Number of milliseconds a token is valid after being authenticated.
* `client_id` - (Optional) Client identifier of the Relying party at the OpenID identity provider. This identifier is typically obtained when the Relying party is registered with the OpenID identity provider. You can specify a regular expression so the AWS AppSync can validate against multiple client identifiers at a time.
* `iat_ttl` - (Optional) Number of milliseconds a token is valid after being issued to a user.

### user_pool_config

The following arguments are supported:

* `default_action` - (Required) The action that you want your GraphQL API to take when a request that uses Amazon Cognito User Pool authentication doesn't match the Amazon Cognito User Pool configuration. Valid: `ALLOW` and `DENY`
* `user_pool_id` - (Required) The user pool ID.
* `app_id_client_regex` - (Optional) A regular expression for validating the incoming Amazon Cognito User Pool app client ID.
* `aws_region` - (Optional) The AWS region in which the user pool was created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - API ID
* `arn` - The ARN
* `uris` - Map of URIs associated with the API. e.g. `uris["GRAPHQL"] = https://ID.appsync-api.REGION.amazonaws.com/graphql`

## Import

AppSync GraphQL API can be imported using the GraphQL API ID, e.g.

```
$ terraform import aws_appsync_graphql_api.example 0123456789
```
