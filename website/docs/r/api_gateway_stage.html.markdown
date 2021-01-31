---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_stage"
description: |-
  Manages an API Gateway Stage.
---

# Resource: aws_api_gateway_stage

Manages an API Gateway Stage. A stage is a named reference to a deployment, which can be done via the [`aws_api_gateway_deployment` resource](api_gateway_deployment.html). Stages can be optionally managed further with the [`aws_api_gateway_base_path_mapping` resource](api_gateway_base_path_mapping.html), [`aws_api_gateway_domain_name` resource](api_gateway_domain_name.html), and [`aws_api_method_settings` resource](api_gateway_method_settings.html). For more information, see the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-stages.html).

## Example Usage

An end-to-end example of a REST API configured with OpenAPI can be found in the [`/examples/api-gateway-rest-api-openapi` directory within the GitHub repository](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples/api-gateway-rest-api-openapi).

```hcl
resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}

resource "aws_api_gateway_method_settings" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = aws_api_gateway_stage.example.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "INFO"
  }
}
```

### Managing the API Logging CloudWatch Log Group

API Gateway provides the ability to [enable CloudWatch API logging](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-logging.html). To manage the CloudWatch Log Group when this feature is enabled, the [`aws_cloudwatch_log_group` resource](/docs/providers/aws/r/cloudwatch_log_group.html) can be used where the name matches the API Gateway naming convention. If the CloudWatch Log Group previously exists, the [`aws_cloudwatch_log_group` resource can be imported into Terraform](/docs/providers/aws/r/cloudwatch_log_group.html#import) as a one time operation and recreation of the environment can occur without import.

-> The below configuration uses [`depends_on`](https://www.terraform.io/docs/configuration/meta-arguments/depends_on.html) to prevent ordering issues with API Gateway automatically creating the log group first and a variable for naming consistency. Other ordering and naming methodologies may be more appropriate for your environment.

```hcl
variable "stage_name" {
  default = "example"
  type    = string
}

resource "aws_api_gateway_rest_api" "example" {
  # ... other configuration ...
}

resource "aws_api_gateway_stage" "example" {
  depends_on = [aws_cloudwatch_log_group.example]

  stage_name = var.stage_name
  # ... other configuration ...
}

resource "aws_cloudwatch_log_group" "example" {
  name              = "API-Gateway-Execution-Logs_${aws_api_gateway_rest_api.example.id}/${var.stage_name}"
  retention_in_days = 7
  # ... potentially other configuration ...
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `stage_name` - (Required) The name of the stage
* `deployment_id` - (Required) The ID of the deployment that the stage points to
* `access_log_settings` - (Optional) Enables access logs for the API stage. Detailed below.
* `cache_cluster_enabled` - (Optional) Specifies whether a cache cluster is enabled for the stage
* `cache_cluster_size` - (Optional) The size of the cache cluster for the stage, if enabled. Allowed values include `0.5`, `1.6`, `6.1`, `13.5`, `28.4`, `58.2`, `118` and `237`.
* `client_certificate_id` - (Optional) The identifier of a client certificate for the stage.
* `description` - (Optional) The description of the stage
* `documentation_version` - (Optional) The version of the associated API documentation
* `variables` - (Optional) A map that defines the stage variables
* `tags` - (Optional) A map of tags to assign to the resource.
* `xray_tracing_enabled` - (Optional) Whether active tracing with X-ray is enabled. Defaults to `false`.

### Nested Blocks

#### `access_log_settings`

* `destination_arn` - (Required) The Amazon Resource Name (ARN) of the CloudWatch Logs log group or Kinesis Data Firehose delivery stream to receive access logs. If you specify a Kinesis Data Firehose delivery stream, the stream name must begin with `amazon-apigateway-`. Automatically removes trailing `:*` if present.
* `format` - (Required) The formatting and values recorded in the logs.
For more information on configuring the log format rules visit the AWS [documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-logging.html)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN)
* `id` - The ID of the stage
* `invoke_url` - The URL to invoke the API pointing to the stage,
  e.g. `https://z4675bid1j.execute-api.eu-west-2.amazonaws.com/prod`
* `execution_arn` - The execution ARN to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn`
  when allowing API Gateway to invoke a Lambda function,
  e.g. `arn:aws:execute-api:eu-west-2:123456789012:z4675bid1j/prod`

## Import

`aws_api_gateway_stage` can be imported using `REST-API-ID/STAGE-NAME`, e.g.

```
$ terraform import aws_api_gateway_stage.example 12345abcde/example
```
