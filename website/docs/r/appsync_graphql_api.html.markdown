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

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A user-supplied name for the GraphqlApi.
* `authentication_type` - (Required) The authentication type. Valid values: `API_KEY`, `AWS_IAM` and `AMAZON_COGNITO_USER_POOLS`
* `user_pool_config` - (Optional) The Amazon Cognito User Pool configuration. See [below](#user_pool_config)

### user_pool_config

The following arguments are supported:

* `aws_region` - (Required) The AWS region in which the user pool was created.
* `default_action` - (Required) The action that you want your GraphQL API to take when a request that uses Amazon Cognito User Pool authentication doesn't match the Amazon Cognito User Pool configuration. Valid: `ALLOW` and `DENY`
* `user_pool_id` - (Required) The user pool ID.
* `app_id_client_regex` - (Optional) A regular expression for validating the incoming Amazon Cognito User Pool app client ID.

## Attributes Reference

The following attributes are exported:

* `id` - API ID
* `arn` - The ARN

## Import

AppSync GraphQL API can be imported using the GraphQL API ID, e.g.

```
$ terraform import aws_appsync_graphql_api.example 0123456789
```
