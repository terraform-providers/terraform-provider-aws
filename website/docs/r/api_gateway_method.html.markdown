---
layout: "aws"
page_title: "AWS: aws_api_gateway_method"
sidebar_current: "docs-aws-resource-api-gateway-method"
description: |-
  Provides a HTTP Method for an API Gateway Resource.
---

# aws_api_gateway_method

Provides a HTTP Method for an API Gateway Resource.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_resource" "MyDemoResource" {
  rest_api_id = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
  parent_id   = "${aws_api_gateway_rest_api.MyDemoAPI.root_resource_id}"
  path_part   = "mydemoresource"
}

resource "aws_api_gateway_method" "MyDemoMethod" {
  rest_api_id   = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
  resource_id   = "${aws_api_gateway_resource.MyDemoResource.id}"
  http_method   = "GET"
  authorization = "NONE"
}
```

## Usage with Cognito User Pool Authorizer
```hcl
variable "cognito_user_pool_name" {}

data "aws_cognito_user_pools" "this" {
  name = "${var.cognito_user_pool_name}"
}

resource "aws_api_gateway_rest_api" "this" {
  name = "with-authorizer"
}

resource "aws_api_gateway_resource" "this" {
  rest_api_id = "${aws_api_gateway_rest_api.this.id}"
  parent_id   = "${aws_api_gateway_rest_api.this.root_resource_id}"
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_authorizer" "this" {
  name          = "CognitoUserPoolAuthorizer"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = "${aws_api_gateway_rest_api.this.id}"
  provider_arns = ["${data.aws_cognito_user_pools.this.arns}"]
}

resource "aws_api_gateway_method" "any" {
  rest_api_id   = "${aws_api_gateway_rest_api.this.id}"
  resource_id   = "${aws_api_gateway_resource.this.id}"
  http_method   = "ANY"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${aws_api_gateway_authorizer.this.id}"

  request_parameters = {
    "method.request.path.proxy" = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `resource_id` - (Required) The API resource ID
* `http_method` - (Required) The HTTP Method (`GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `OPTIONS`, `ANY`)
* `authorization` - (Required) The type of authorization used for the method (`NONE`, `CUSTOM`, `AWS_IAM`, `COGNITO_USER_POOLS`)
* `authorizer_id` - (Optional) The authorizer id to be used when the authorization is `CUSTOM` or `COGNITO_USER_POOLS`
* `authorization_scopes` - (Optional) The authorization scopes used when the authorization is `COGNITO_USER_POOLS`
* `api_key_required` - (Optional) Specify if the method requires an API key
* `request_models` - (Optional) A map of the API models used for the request's content type
  where key is the content type (e.g. `application/json`)
  and value is either `Error`, `Empty` (built-in models) or `aws_api_gateway_model`'s `name`.
* `request_validator_id` - (Optional) The ID of a `aws_api_gateway_request_validator`
* `request_parameters` - (Optional) A map of request query string parameters and headers that should be passed to the integration.
  For example:
```hcl
request_parameters = {
  "method.request.header.X-Some-Header"         = true
  "method.request.querystring.some-query-param" = true
}
```
would define that the header `X-Some-Header` and the query string `some-query-param` must be provided on the request, or
* `request_parameters_in_json` - **Deprecated**, use `request_parameters` instead.

## Import

`aws_api_gateway_method` can be imported using `REST-API-ID/RESOURCE-ID/HTTP-METHOD`, e.g.

```
$ terraform import aws_api_gateway_method.example 12345abcde/67890fghij/GET
```
