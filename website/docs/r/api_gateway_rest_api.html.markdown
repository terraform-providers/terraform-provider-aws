---
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api"
sidebar_current: "docs-aws-resource-api-gateway-rest-api"
description: |-
  Provides an API Gateway REST API.
---

# aws_api_gateway_rest_api

Provides an API Gateway REST API.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the REST API
* `description` - (Optional) The description of the REST API
* `binary_media_types` - (Optional) The list of binary media types supported by the RestApi. By default, the RestApi supports only UTF-8-encoded text payloads.
* `minimum_compression_size` - (Optional) Minimum response size to compress for the REST API. Integer between -1 and 10485760 (10MB). Setting a value greater than -1 will enable compression, -1 disables compression (default).
* `body` - (Optional) An OpenAPI specification that defines the set of routes and integrations to create as part of the REST API.
* `policy` - (Optional) JSON formatted policy document that controls access to the API Gateway

__Note__: If the `body` argument is provided, the OpenAPI specification will be used to configure the resources, methods and integrations for the Rest API. If this argument is provided, the following resources should not be managed as separate ones, as updates may cause manual resource updates to be overwritten:

* `aws_api_gateway_resource`
* `aws_api_gateway_method`
* `aws_api_gateway_method_response`
* `aws_api_gateway_method_settings`
* `aws_api_gateway_integration`
* `aws_api_gateway_integration_response`
* `aws_api_gateway_gateway_response`
* `aws_api_gateway_model`

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the REST API
* `root_resource_id` - The resource ID of the REST API's root
* `created_date` - The creation date of the REST API
* `execution_arn` - The execution ARN part to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn`
  when allowing API Gateway to invoke a Lambda function,
  e.g. `arn:aws:execute-api:eu-west-2:123456789012:z4675bid1j`, which can be concatenated with allowed stage, method and resource path.
