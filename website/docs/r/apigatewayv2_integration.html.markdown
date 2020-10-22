---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_integration"
description: |-
  Manages an Amazon API Gateway Version 2 integration.
---

# Resource: aws_apigatewayv2_integration

Manages an Amazon API Gateway Version 2 integration.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_integration" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  integration_type = "MOCK"
}
```

### Lambda Integration

```hcl
resource "aws_lambda_function" "example" {
  filename      = "example.zip"
  function_name = "Example"
  role          = aws_iam_role.example.arn
  handler       = "index.handler"
  runtime       = "nodejs12.x"
}

resource "aws_apigatewayv2_integration" "example" {
  api_id           = aws_apigatewayv2_api.example.id
  integration_type = "AWS"

  connection_type           = "INTERNET"
  content_handling_strategy = "CONVERT_TO_TEXT"
  description               = "Lambda example"
  integration_method        = "POST"
  integration_uri           = aws_lambda_function.example.invoke_arn
  passthrough_behavior      = "WHEN_NO_MATCH"
}
```

### AWS Service Integration

```hcl
resource "aws_apigatewayv2_integration" "example" {
  api_id              = aws_apigatewayv2_api.example.id
  credentials_arn     = aws_iam_role.example.arn
  description         = "SQS example"
  integration_type    = "AWS_PROXY"
  integration_subtype = "SQS-SendMessage"

  request_parameters = {
    "QueueUrl"    = "$request.header.queueUrl"
    "MessageBody" = "$request.body.message"
  }
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `integration_type` - (Required) The integration type of an integration.
Valid values: `AWS`, `AWS_PROXY`, `HTTP`, `HTTP_PROXY`, `MOCK`.
* `connection_id` - (Optional) The ID of the VPC link for a private integration. Supported only for HTTP APIs.
* `connection_type` - (Optional) The type of the network connection to the integration endpoint. Valid values: `INTERNET`, `VPC_LINK`. Default is `INTERNET`.
* `content_handling_strategy` - (Optional) How to handle response payload content type conversions. Valid values: `CONVERT_TO_BINARY`, `CONVERT_TO_TEXT`. Supported only for WebSocket APIs.
* `credentials_arn` - (Optional) The credentials required for the integration, if any.
* `description` - (Optional) The description of the integration.
* `integration_method` - (Optional) The integration's HTTP method. Must be specified if `integration_type` is not `MOCK`.
* `integration_subtype` - (Optional) Specifies the AWS service action to invoke. Supported only for HTTP APIs when `integration_type` is `AWS_PROXY`. See the [AWS service integration reference](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-aws-services-reference.html) documentation for supported values.
* `integration_uri` - (Optional) The URI of the Lambda function for a Lambda proxy integration, when `integration_type` is `AWS_PROXY`.
For an `HTTP` integration, specify a fully-qualified URL. For an HTTP API private integration, specify the ARN of an Application Load Balancer listener, Network Load Balancer listener, or AWS Cloud Map service.
* `passthrough_behavior` - (Optional) The pass-through behavior for incoming requests based on the Content-Type header in the request, and the available mapping templates specified as the `request_templates` attribute.
Valid values: `WHEN_NO_MATCH`, `WHEN_NO_TEMPLATES`, `NEVER`. Default is `WHEN_NO_MATCH`. Supported only for WebSocket APIs.
* `payload_format_version` - (Optional) The [format of the payload](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-integrations-lambda.html#http-api-develop-integrations-lambda.proxy-format) sent to an integration. Valid values: `1.0`, `2.0`. Default is `1.0`.
* `request_parameters` - (Optional) A key-value map specifying request parameters that are passed from the method request to the backend.
Supported only for WebSocket APIs.
* `request_templates` - (Optional) A map of Velocity templates that are applied on the request payload based on the value of the Content-Type header sent by the client. Supported only for WebSocket APIs.
* `template_selection_expression` - (Optional) The [template selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-template-selection-expressions) for the integration.
* `timeout_milliseconds` - (Optional) Custom timeout between 50 and 29,000 milliseconds. The default value is 29,000 milliseconds or 29 seconds.
* `tls_config` - (Optional) The TLS configuration for a private integration. Supported only for HTTP APIs.

The `tls_config` object supports the following:

* `server_name_to_verify` - (Optional) If you specify a server name, API Gateway uses it to verify the hostname on the integration's certificate. The server name is also included in the TLS handshake to support Server Name Indication (SNI) or virtual hosting.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The integration identifier.
* `integration_response_selection_expression` - The [integration response selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-integration-response-selection-expressions) for the integration.

## Import

`aws_apigatewayv2_integration` can be imported by using the API identifier and integration identifier, e.g.

```
$ terraform import aws_apigatewayv2_integration.example aabbccddee/1122334
```

-> **Note:** The API Gateway managed integration created as part of [_quick_create_](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-basic-concept.html#apigateway-definition-quick-create) cannot be imported.
