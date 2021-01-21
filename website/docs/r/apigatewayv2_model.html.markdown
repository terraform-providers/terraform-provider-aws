---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_model"
description: |-
  Manages an Amazon API Gateway Version 2 model.
---

# Resource: aws_apigatewayv2_model

Manages an Amazon API Gateway Version 2 [model](https://docs.aws.amazon.com/apigateway/latest/developerguide/models-mappings.html#models-mappings-models).

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_model" "example" {
  api_id       = aws_apigatewayv2_api.example.id
  content_type = "application/json"
  name         = "example"

  schema = <<EOF
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "ExampleModel",
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  }
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `content_type` - (Required)  The content-type for the model, for example, `application/json`. Must be between 1 and 256 characters in length.
* `name` - (Required) The name of the model. Must be alphanumeric. Must be between 1 and 128 characters in length.
* `schema` - (Required) The schema for the model. This should be a [JSON schema draft 4](https://tools.ietf.org/html/draft-zyp-json-schema-04) model. Must be less than or equal to 32768 characters in length.
* `description` - (Optional) The description of the model. Must be between 1 and 128 characters in length.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The model identifier.

## Import

`aws_apigatewayv2_model` can be imported by using the API identifier and model identifier, e.g.

```
$ terraform import aws_apigatewayv2_model.example aabbccddee/1122334
```
