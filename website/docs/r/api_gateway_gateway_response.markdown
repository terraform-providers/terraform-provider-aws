---
layout: "aws"
page_title: "AWS: aws_api_gateway_gateway_response"
sidebar_current: "docs-aws-resource-api-gateway-gateway-response"
description: |-
  Provides an API Gateway Gateway Response for a REST API Gateway.
---

# aws_api_gateway_gateway_response

Provides an API Gateway Gateway Response for a REST API Gateway.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "main" {
  name = "MyDemoAPI"
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = "${aws_api_gateway_rest_api.main.id}"
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.Authorization" = "'Basic'"
  }
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The string identifier of the associated REST API.
* `response_type` - (Required) The response type of the associated GatewayResponse.
* `status_code` - (Optional) The HTTP status code of the Gateway Response.
* `response_parameters` - (Optional) A map specifying the templates used to transform the response body.
* `response_templates` - (Optional) A map specifying the parameters (paths, query strings and headers) of the Gateway Response.

## Import

`aws_api_gateway_gateway_response` can be imported using `REST-API-ID/RESPONSE-TYPE`, e.g.

```
$ terraform import aws_api_gateway_gateway_response.example 12345abcde/UNAUTHORIZED
```
