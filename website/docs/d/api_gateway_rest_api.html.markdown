---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_rest_api"
description: |-
  Get information on a API Gateway REST API
---

# Data Source: aws_api_gateway_rest_api

Use this data source to get the id and root_resource_id of a REST API in
API Gateway. To fetch the REST API you must provide a name to match against.
As there is no unique name constraint on REST APIs this data source will
error if there is more than one match.

## Example Usage

```hcl
data "aws_api_gateway_rest_api" "my_rest_api" {
  name = "my-rest-api"
}
```

## Argument Reference

* `name` - (Required) The name of the REST API to look up. If no REST API is found with this name, an error will be returned. If multiple REST APIs are found with this name, an error will be returned.

## Attributes Reference

* `api_key_source` - The source of the API key for requests.
* `arn` - The ARN of the REST API.
* `binary_media_types` - The list of binary media types supported by the REST API.
* `description` - The description of the REST API.
* `endpoint_configuration` - The endpoint configuration of this RestApi showing the endpoint types of the API.
* `execution_arn` - The execution ARN part to be used in [`lambda_permission`](/docs/providers/aws/r/lambda_permission.html)'s `source_arn` when allowing API Gateway to invoke a Lambda function, e.g. `arn:aws:execute-api:eu-west-2:123456789012:z4675bid1j`, which can be concatenated with allowed stage, method and resource path.
* `id` - Set to the ID of the found REST API.
* `minimum_compression_size` - Minimum response size to compress for the REST API.
* `policy` - JSON formatted policy document that controls access to the API Gateway.
* `root_resource_id` - Set to the ID of the API Gateway Resource on the found REST API where the route matches '/'.
* `tags` - Key-value map of resource tags.
