---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_route"
description: |-
  Manages an Amazon API Gateway Version 2 route.
---

# Resource: aws_apigatewayv2_route

Manages an Amazon API Gateway Version 2 route.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_route" "example" {
  api_id    = "${aws_apigatewayv2_api.example.id}"
  route_key = "$default"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `route_key` - (Required) The route key for the route.
* `api_key_required` - (Optional) Boolean whether an API key is required for the route. Defaults to `false`.
* `authorization_scopes` - (Optional) The authorization scopes supported by this route. The scopes are used with a JWT authorizer to authorize the method invocation.
* `authorization_type` - (Optional) The authorization type for the route.
For WebSocket APIs, valid values are `NONE` for open access, `AWS_IAM` for using AWS IAM permissions, and `CUSTOM` for using a Lambda authorizer.
For HTTP APIs, valid values are `NONE` for open access, or `JWT` for using JSON Web Tokens.
Defaults to `NONE`.
* `authorizer_id` - (Optional) The identifier of the [`aws_apigatewayv2_authorizer`](/docs/providers/aws/r/apigatewayv2_authorizer.html) resource to be associated with this route, if the authorizationType is `CUSTOM`.
* `model_selection_expression` - (Optional) The [model selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-model-selection-expressions) for the route.
* `operation_name` - (Optional) The operation name for the route.
* `request_models` - (Optional) The request models for the route.
* `route_response_selection_expression` - (Optional) The [route response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-response-selection-expressions) for the route.
* `target` - (Optional) The target for the route.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The route identifier.

## Import

`aws_apigatewayv2_route` can be imported by using the API identifier and route identifier, e.g.

```
$ terraform import aws_apigatewayv2_route.example aabbccddee/1122334
```

-> **Note:** The API Gateway managed route created as part of [_quick_create_](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-basic-concept.html#apigateway-definition-quick-create) cannot be imported.
