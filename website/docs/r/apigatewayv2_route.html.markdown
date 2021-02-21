---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_route"
description: |-
  Manages an Amazon API Gateway Version 2 route.
---

# Resource: aws_apigatewayv2_route

Manages an Amazon API Gateway Version 2 route.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/welcome.html) for [WebSocket](https://docs.aws.amazon.com/apigateway/latest/developerguide/websocket-api-develop-routes.html) and [HTTP](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-routes.html) APIs.

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_api" "example" {
  name                       = "example-websocket-api"
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "$default"
}
```

### HTTP Proxy Integration

```hcl
resource "aws_apigatewayv2_api" "example" {
  name          = "example-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  integration_type = "HTTP_PROXY"

  integration_method = "ANY"
  integration_uri    = "https://example.com/{proxy}"
}

resource "aws_apigatewayv2_route" "example" {
  api_id    = aws_apigatewayv2_api.example.id
  route_key = "ANY /example/{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `route_key` - (Required) The route key for the route. For HTTP APIs, the route key can be either `$default`, or a combination of an HTTP method and resource path, for example, `GET /pets`.
* `api_key_required` - (Optional) Boolean whether an API key is required for the route. Defaults to `false`. Supported only for WebSocket APIs.
* `authorization_scopes` - (Optional) The authorization scopes supported by this route. The scopes are used with a JWT authorizer to authorize the method invocation.
* `authorization_type` - (Optional) The authorization type for the route.
For WebSocket APIs, valid values are `NONE` for open access, `AWS_IAM` for using AWS IAM permissions, and `CUSTOM` for using a Lambda authorizer.
For HTTP APIs, valid values are `NONE` for open access, `JWT` for using JSON Web Tokens, `AWS_IAM` for using AWS IAM permissions, and `CUSTOM` for using a Lambda authorizer.
Defaults to `NONE`.
* `authorizer_id` - (Optional) The identifier of the [`aws_apigatewayv2_authorizer`](apigatewayv2_authorizer.html) resource to be associated with this route.
* `model_selection_expression` - (Optional) The [model selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-model-selection-expressions) for the route. Supported only for WebSocket APIs.
* `operation_name` - (Optional) The operation name for the route. Must be between 1 and 64 characters in length.
* `request_models` - (Optional) The request models for the route. Supported only for WebSocket APIs.
* `route_response_selection_expression` - (Optional) The [route response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-response-selection-expressions) for the route. Supported only for WebSocket APIs.
* `target` - (Optional) The target for the route, of the form `integrations/`*`IntegrationID`*, where *`IntegrationID`* is the identifier of an [`aws_apigatewayv2_integration`](apigatewayv2_integration.html) resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The route identifier.

## Import

`aws_apigatewayv2_route` can be imported by using the API identifier and route identifier, e.g.

```
$ terraform import aws_apigatewayv2_route.example aabbccddee/1122334
```

-> **Note:** The API Gateway managed route created as part of [_quick_create_](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-basic-concept.html#apigateway-definition-quick-create) cannot be imported.
