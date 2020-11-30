---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_stage"
description: |-
  Manages an Amazon API Gateway Version 2 stage.
---

# Resource: aws_apigatewayv2_stage

Manages an Amazon API Gateway Version 2 stage.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api.html).

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_stage" "example" {
  api_id = aws_apigatewayv2_api.example.id
  name   = "example-stage"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `name` - (Required) The name of the stage. Must be between 1 and 128 characters in length.
* `access_log_settings` - (Optional) Settings for logging access in this stage.
Use the [`aws_api_gateway_account`](/docs/providers/aws/r/api_gateway_account.html) resource to configure [permissions for CloudWatch Logging](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-logging.html#set-up-access-logging-permissions).
* `auto_deploy` - (Optional) Whether updates to an API automatically trigger a new deployment. Defaults to `false`. Applicable for HTTP APIs.
* `client_certificate_id` - (Optional) The identifier of a client certificate for the stage. Use the [`aws_api_gateway_client_certificate`](/docs/providers/aws/r/api_gateway_client_certificate.html) resource to configure a client certificate.
Supported only for WebSocket APIs.
* `default_route_settings` - (Optional) The default route settings for the stage.
* `deployment_id` - (Optional) The deployment identifier of the stage. Use the [`aws_apigatewayv2_deployment`](/docs/providers/aws/r/apigatewayv2_deployment.html) resource to configure a deployment.
* `description` - (Optional) The description for the stage. Must be less than or equal to 1024 characters in length.
* `route_settings` - (Optional) Route settings for the stage.
* `stage_variables` - (Optional) A map that defines the stage variables for the stage.
* `tags` - (Optional) A map of tags to assign to the stage.

The `access_log_settings` object supports the following:

* `destination_arn` - (Required) The ARN of the CloudWatch Logs log group to receive access logs. Any trailing `:*` is trimmed from the ARN.
* `format` - (Required) A single line [format](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-logging.html#apigateway-cloudwatch-log-formats) of the access logs of data, as specified by [selected $context variables](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-logging.html).

The `default_route_settings` object supports the following:

* `data_trace_enabled` - (Optional) Whether data trace logging is enabled for the default route. Affects the log entries pushed to Amazon CloudWatch Logs.
Defaults to `false`. Supported only for WebSocket APIs.
* `detailed_metrics_enabled` - (Optional) Whether detailed metrics are enabled for the default route. Defaults to `false`.
* `logging_level` - (Optional) The logging level for the default route. Affects the log entries pushed to Amazon CloudWatch Logs.
Valid values: `ERROR`, `INFO`, `OFF`. Defaults to `OFF`. Supported only for WebSocket APIs. Terraform will only perform drift detection of its value when present in a configuration.
* `throttling_burst_limit` - (Optional) The throttling burst limit for the default route.
* `throttling_rate_limit` - (Optional) The throttling rate limit for the default route.

The `route_settings` object supports the following:

* `route_key` - (Required) Route key.
* `data_trace_enabled` - (Optional) Whether data trace logging is enabled for the route. Affects the log entries pushed to Amazon CloudWatch Logs.
Defaults to `false`. Supported only for WebSocket APIs.
* `detailed_metrics_enabled` - (Optional) Whether detailed metrics are enabled for the route. Defaults to `false`.
* `logging_level` - (Optional) The logging level for the route. Affects the log entries pushed to Amazon CloudWatch Logs.
Valid values: `ERROR`, `INFO`, `OFF`. Defaults to `OFF`. Supported only for WebSocket APIs. Terraform will only perform drift detection of its value when present in a configuration.
* `throttling_burst_limit` - (Optional) The throttling burst limit for the route.
* `throttling_rate_limit` - (Optional) The throttling rate limit for the route.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The stage identifier.
* `arn` - The ARN of the stage.
* `execution_arn` - The ARN prefix to be used in an [`aws_lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn` attribute.
For WebSocket APIs this attribute can additionally be used in an [`aws_iam_policy`](/docs/providers/aws/r/iam_policy.html) to authorize access to the [`@connections` API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-how-to-call-websocket-api-connections.html).
See the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-control-access-iam.html) for details.
* `invoke_url` - The URL to invoke the API pointing to the stage,
  e.g. `wss://z4675bid1j.execute-api.eu-west-2.amazonaws.com/example-stage`, or `https://z4675bid1j.execute-api.eu-west-2.amazonaws.com/`

## Import

`aws_apigatewayv2_stage` can be imported by using the API identifier and stage name, e.g.

```
$ terraform import aws_apigatewayv2_stage.example aabbccddee/example-stage
```

-> **Note:** The API Gateway managed stage created as part of [_quick_create_](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-basic-concept.html#apigateway-definition-quick-create) cannot be imported.
