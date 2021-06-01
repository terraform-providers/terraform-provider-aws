---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api"
description: |-
  Provides an API Gateway REST API.
---

# Resource: aws_api_gateway_rest_api

Provides an API Gateway REST API.

-> **Note:** Amazon API Gateway Version 1 resources are used for creating and deploying REST APIs. To create and deploy WebSocket and HTTP APIs, use Amazon API Gateway Version 2 [resources](/docs/providers/aws/r/apigatewayv2_api.html).

## Example Usage

### Basic

```hcl
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}
```

### Regional Endpoint Type

```hcl
resource "aws_api_gateway_rest_api" "example" {
  name = "regional-example"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the REST API. If importing an OpenAPI specification via the `body` argument, this corresponds to the `info.title` field. If the argument value is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `description` - (Optional) Description of the REST API. If importing an OpenAPI specification via the `body` argument, this corresponds to the `info.description` field. If the argument value is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `endpoint_configuration` - (Optional) Configuration block defining API endpoint configuration including endpoint type. Defined below.
* `binary_media_types` - (Optional) List of binary media types supported by the REST API. By default, the REST API supports only UTF-8-encoded text payloads. If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-binary-media-types` extension](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-binary-media-types.html). If the argument value is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `minimum_compression_size` - (Optional) Minimum response size to compress for the REST API. Integer between `-1` and `10485760` (10MB). Setting a value greater than `-1` will enable compression, `-1` disables compression (default). If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-minimum-compression-size` extension](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-openapi-minimum-compression-size.html). If the argument value (_except_ `-1`) is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `body` - (Optional) OpenAPI specification that defines the set of routes and integrations to create as part of the REST API. This configuration, and any updates to it, will replace all REST API configuration except values overridden in this resource configuration and other resource updates applied after this resource but before any `aws_api_gateway_deployment` creation. More information about REST API OpenAPI support can be found in the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-import-api.html).
* `parameters` - (Optional) Map of customizations for importing the specification in the `body` argument. For example, to exclude DocumentationParts from an imported API, set `ignore` equal to `documentation`. Additional documentation, including other parameters such as `basepath`, can be found in the [API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-import-api.html).
* `policy` - (Optional) JSON formatted policy document that controls access to the API Gateway. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Terraform will only perform drift detection of its value when present in a configuration. It is recommended to use the [`aws_api_gateway_rest_api_policy` resource](/docs/providers/aws/r/api_gateway_rest_api_policy.html) instead. If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-policy` extension](https://docs.aws.amazon.com/apigateway/latest/developerguide/openapi-extensions-policy.html). If the argument value is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `api_key_source` - (Optional) Source of the API key for requests. Valid values are `HEADER` (default) and `AUTHORIZER`. If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-api-key-source` extension](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-api-key-source.html). If the argument value is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `disable_execute_api_endpoint` - (Optional) Specifies whether clients can invoke your API by using the default execute-api endpoint. By default, clients can invoke your API with the default https://{api_id}.execute-api.{region}.amazonaws.com endpoint. To require that clients use a custom domain name to invoke your API, disable the default endpoint. Defaults to `false`. If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-endpoint-configuration` extension `disableExecuteApiEndpoint` property](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-endpoint-configuration.html). If the argument value is `true` and is different than the OpenAPI value, the argument value will override the OpenAPI value.
* `tags` - (Optional) Key-value map of resource tags

__Note__: If the `body` argument is provided, the OpenAPI specification will be used to configure the resources, methods and integrations for the Rest API. If this argument is provided, the following resources should not be managed as separate ones, as updates may cause manual resource updates to be overwritten:

* `aws_api_gateway_resource`
* `aws_api_gateway_method`
* `aws_api_gateway_method_response`
* `aws_api_gateway_method_settings`
* `aws_api_gateway_integration`
* `aws_api_gateway_integration_response`
* `aws_api_gateway_gateway_response`
* `aws_api_gateway_model`

### endpoint_configuration

* `types` - (Required) A list of endpoint types. This resource currently only supports managing a single value. Valid values: `EDGE`, `REGIONAL` or `PRIVATE`. If unspecified, defaults to `EDGE`. Must be declared as `REGIONAL` in non-Commercial partitions. Refer to the [documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/create-regional-api.html) for more information on the difference between edge-optimized and regional APIs.
* `vpc_endpoint_ids` - (Optional) Set of VPC Endpoint identifiers. It is only supported for `PRIVATE` endpoint type. If importing an OpenAPI specification via the `body` argument, this corresponds to the [`x-amazon-apigateway-endpoint-configuration` extension `vpcEndpointIds` property](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-endpoint-configuration.html). If the argument value is provided and is different than the OpenAPI value, the argument value will override the OpenAPI value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the REST API
* `root_resource_id` - The resource ID of the REST API's root
* `created_date` - The creation date of the REST API
* `execution_arn` - The execution ARN part to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn`
  when allowing API Gateway to invoke a Lambda function,
  e.g. `arn:aws:execute-api:eu-west-2:123456789012:z4675bid1j`, which can be concatenated with allowed stage, method and resource path.
* `arn` - Amazon Resource Name (ARN)

## Import

`aws_api_gateway_rest_api` can be imported by using the REST API ID, e.g.

```
$ terraform import aws_api_gateway_rest_api.example 12345abcde
```

~> **NOTE:** Resource import does not currently support the `body` attribute.
