---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api"
description: |-
  Manages an Amazon API Gateway Version 2 API.
---

# Resource: aws_apigatewayv2_api

Manages an Amazon API Gateway Version 2 API.

-> **Note:** Amazon API Gateway Version 2 resources are used for creating and deploying WebSocket and HTTP APIs. To create and deploy REST APIs, use Amazon API Gateway Version 1 [resources](https://www.terraform.io/docs/providers/aws/r/api_gateway_rest_api.html).

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

* `name` - (Required) The name of the API.
* `protocol_type` - (Required) The API protocol. Valid values: `HTTP`, `WEBSOCKET`.
* `api_key_selection_expression` - (Optional) An [API key selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-apikey-selection-expressions).
Valid values: `$context.authorizer.usageIdentifierKey`, `$request.header.x-api-key`. Defaults to `$request.header.x-api-key`.
Applicable for WebSocket APIs.
* `description` - (Optional) The description of the API.
* `route_selection_expression` - (Optional) The [route selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-route-selection-expressions) for the API.
Defaults to `$request.method $request.path`.
* `tags` - (Optional) A mapping of tags to assign to the API.
* `version` - (Optional) A version identifier for the API.

## Attribute Reference

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
