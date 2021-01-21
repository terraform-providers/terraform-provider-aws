---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api"
description: |-
  Manages an Amazon API Gateway Version 2 API.
---

# Resource: aws_apigatewayv2_api

Manages an Amazon API Gateway Version 2 API.

-> **Note:** Amazon API Gateway Version 2 resources are used for creating and deploying WebSocket and HTTP APIs. To create and deploy REST APIs, use Amazon API Gateway Version 1 [resources](/docs/providers/aws/r/api_gateway_rest_api.html).

## Example Usage

### Basic WebSocket API

```hcl
resource "aws_apigatewayv2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
```

### Basic HTTP API

```hcl
resource "aws_apigatewayv2_api" "example" {
  name          = "example-http-api"
  protocol_type = "HTTP"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the API. Must be less than or equal to 128 characters in length.
* `protocol_type` - (Required) The API protocol. Valid values: `HTTP`, `WEBSOCKET`.
* `api_key_selection_expression` - (Optional) An [API key selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-apikey-selection-expressions).
Valid values: `$context.authorizer.usageIdentifierKey`, `$request.header.x-api-key`. Defaults to `$request.header.x-api-key`.
Applicable for WebSocket APIs.
* `cors_configuration` - (Optional) The cross-origin resource sharing (CORS) [configuration](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-cors.html). Applicable for HTTP APIs.
* `credentials_arn` - (Optional) Part of _quick create_. Specifies any credentials required for the integration. Applicable for HTTP APIs.
* `description` - (Optional) The description of the API. Must be less than or equal to 1024 characters in length.
* `disable_execute_api_endpoint` - (Optional) Whether clients can invoke the API by using the default `execute-api` endpoint.
By default, clients can invoke the API with the default `{api_id}.execute-api.{region}.amazonaws.com endpoint`.
To require that clients use a custom domain name to invoke the API, disable the default endpoint.
* `route_key` - (Optional) Part of _quick create_. Specifies any [route key](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-routes.html). Applicable for HTTP APIs.
* `route_selection_expression` - (Optional) The [route selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-selection-expressions) for the API.
Defaults to `$request.method $request.path`.
* `tags` - (Optional) A map of tags to assign to the API.
* `target` - (Optional) Part of _quick create_. Quick create produces an API with an integration, a default catch-all route, and a default stage which is configured to automatically deploy changes.
For HTTP integrations, specify a fully qualified URL. For Lambda integrations, specify a function ARN.
The type of the integration will be `HTTP_PROXY` or `AWS_PROXY`, respectively. Applicable for HTTP APIs.
* `body` - (Optional) An OpenAPI specification that defines the set of routes and integrations to create as part of the HTTP APIs. Supported only for HTTP APIs.
* `version` - (Optional) A version identifier for the API. Must be between 1 and 64 characters in length.

__Note__: If the `body` argument is provided, the OpenAPI specification will be used to configure the integrations and route for the HTTP API. If this argument is provided, the following resources should not be managed as separate ones, as updates may cause manual resource updates to be overwritten:

* `aws_apigatewayv2_integration`
* `aws_apigatewayv2_route`

Further more, the `name`, `description`, `cors_configuration`, `tags` and `version` fields should be specified in the Terraform configuration and the values will override any values specified in the OpenAPI document.

The `cors_configuration` object supports the following:

* `allow_credentials` - (Optional) Whether credentials are included in the CORS request.
* `allow_headers` - (Optional) The set of allowed HTTP headers.
* `allow_methods` - (Optional) The set of allowed HTTP methods.
* `allow_origins` - (Optional) The set of allowed origins.
* `expose_headers` - (Optional) The set of exposed HTTP headers.
* `max_age` - (Optional) The number of seconds that the browser should cache preflight request results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The API identifier.
* `api_endpoint` - The URI of the API, of the form `{api-id}.execute-api.{region}.amazonaws.com`.
* `arn` - The ARN of the API.
* `execution_arn` - The ARN prefix to be used in an [`aws_lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn` attribute
or in an [`aws_iam_policy`](/docs/providers/aws/r/iam_policy.html) to authorize access to the [`@connections` API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-how-to-call-websocket-api-connections.html).
See the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-control-access-iam.html) for details.

## Import

`aws_apigatewayv2_api` can be imported by using the API identifier, e.g.

```
$ terraform import aws_apigatewayv2_api.example aabbccddee
```
