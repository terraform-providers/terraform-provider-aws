---
layout: "aws"
page_title: "AWS: aws_api_gateway_request_validator"
sidebar_current: "docs-aws-resource-api-gateway-request-validator"
description: |-
  Manages an API Gateway Request Validator.
---

# aws_api_gateway_request_validator

Manages an API Gateway Request Validator.

## Example Usage

```hcl
resource "aws_api_gateway_request_validator" "example" {
  name                        = "example"
  rest_api_id                 = "${aws_api_gateway_rest_api.example.id}"
  validate_request_body       = true
  validate_request_parameters = true
}
```

## Argument Reference

The following argument is supported:

* `name` - (Required) The name of the request validator
* `rest_api_id` - (Required) The ID of the associated Rest API
* `validate_request_body` - (Optional) Boolean whether to validate request body. Defaults to `false`.
* `validate_request_parameters` - (Optional) Boolean whether to validate request parameters. Defaults to `false`.

## Attribute Reference

The following attribute is exported in addition to the arguments listed above:

* `id` - The unique ID of the request validator

## Import

`aws_api_gateway_request_validator` can be imported using `REST-API-ID/REQUEST-VALIDATOR-ID`, e.g.

```
$ terraform import aws_api_gateway_request_validator.example 12345abcde/67890fghij
```
