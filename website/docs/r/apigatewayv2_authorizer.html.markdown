---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_authorizer"
description: |-
  Manages an Amazon API Gateway Version 2 authorizer.
---

# Resource: aws_apigatewayv2_authorizer

Manages an Amazon API Gateway Version 2 authorizer.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic WebSocket API

```hcl
resource "aws_apigatewayv2_authorizer" "example" {
  api_id           = "${aws_apigatewayv2_api.example.id}"
  authorizer_type  = "REQUEST"
  authorizer_uri   = "${aws_lambda_function.example.invoke_arn}"
  identity_sources = ["route.request.header.Auth"]
  name             = "example-authorizer"
}
```

### Basic HTTP API

```hcl
resource "aws_apigatewayv2_authorizer" "example" {
  api_id           = "${aws_apigatewayv2_api.example.id}"
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "example-authorizer"

  jwt_configuration {
    audience = ["example"]
    issuer   = "https://${aws_cognito_user_pool.example.endpoint}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `authorizer_type` - (Required) The authorizer type. Valid values: `JWT`, `REQUEST`.
For WebSocket APIs, specify `REQUEST` for a Lambda function using incoming request parameters.
 For HTTP APIs, specify `JWT` to use JSON Web Tokens.
* `identity_sources` - (Required) The identity sources for which authorization is requested.
For `REQUEST` authorizers the value is a list of one or more mapping expressions of the specified request parameters.
For `JWT` authorizers the single entry specifies where to extract the JSON Web Token (JWT) from inbound requests.
* `name` - (Required) The name of the authorizer.
* `authorizer_credentials_arn` - (Optional) The required credentials as an IAM role for API Gateway to invoke the authorizer.
Supported only for `REQUEST` authorizers.
* `authorizer_uri` - (Optional) The authorizer's Uniform Resource Identifier (URI).
For `REQUEST` authorizers this must be a well-formed Lambda function URI, such as the `invoke_arn` attribute of the [`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html) resource.
Supported only for `REQUEST` authorizers.
* `jwt_configuration` - (Optional) The configuration of a JWT authorizer. Required for the `JWT` authorizer type.
Supported only for HTTP APIs.

The `jwt_configuration` object supports the following:

* `audience` - (Optional) A list of the intended recipients of the JWT. A valid JWT must provide an aud that matches at least one entry in this list.
* `issuer` - (Optional) The base domain of the identity provider that issues JSON Web Tokens, such as the `endpoint` attribute of the [`aws_cognito_user_pool`](/docs/providers/aws/r/cognito_user_pool.html) resource.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The authorizer identifier.

## Import

`aws_apigatewayv2_authorizer` can be imported by using the API identifier and authorizer identifier, e.g.

```
$ terraform import aws_apigatewayv2_authorizer.example aabbccddee/1122334
```
