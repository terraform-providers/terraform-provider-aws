---
layout: "aws"
page_title: "AWS: aws_api_gateway_model"
sidebar_current: "docs-aws-resource-api-gateway-model"
description: |-
  Provides a Model for a API Gateway.
---

# aws_api_gateway_model

Provides a Model for a API Gateway.

## Example Usage

```hcl
resource "aws_api_gateway_rest_api" "MyDemoAPI" {
  name        = "MyDemoAPI"
  description = "This is my API for demonstration purposes"
}

resource "aws_api_gateway_model" "MyDemoModel" {
  rest_api_id  = "${aws_api_gateway_rest_api.MyDemoAPI.id}"
  name         = "user"
  description  = "a JSON schema"
  content_type = "application/json"

  schema = <<EOF
{
  "type": "object"
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `rest_api_id` - (Required) The ID of the associated REST API
* `name` - (Required) The name of the model
* `description` - (Optional) The description of the model
* `content_type` - (Required) The content type of the model
* `schema` - (Required) The schema of the model in a JSON form

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the model

## Import

`aws_api_gateway_model` can be imported using `REST-API-ID/NAME`, e.g.

```
$ terraform import aws_api_gateway_model.example 12345abcde/example
```
