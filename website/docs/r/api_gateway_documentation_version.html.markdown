---
layout: "aws"
page_title: "AWS: aws_api_gateway_documentation_version"
sidebar_current: "docs-aws-resource-api-gateway-documentation-version"
description: |-
  Provides a resource to manage an API Gateway Documentation Version.
---

# aws_api_gateway_documentation_version

Provides a resource to manage an API Gateway Documentation Version.

## Example Usage

```hcl
resource "aws_api_gateway_documentation_version" "example" {
  version     = "example_version"
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  description = "Example description"
  depends_on  = ["aws_api_gateway_documentation_part.example"]
}

resource "aws_api_gateway_rest_api" "example" {
  name = "example_api"
}

resource "aws_api_gateway_documentation_part" "example" {
  location {
    type = "API"
  }

  properties  = "{\"description\":\"Example\"}"
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
}
```

## Argument Reference

The following argument is supported:

* `version` - (Required) The version identifier of the API documentation snapshot.
* `rest_api_id` - (Required) The ID of the associated Rest API
* `description` - (Optional) The description of the API documentation version.

## Attribute Reference

The arguments listed above are all exported as attributes.

## Import

API Gateway documentation versions can be imported using `REST-API-ID/VERSION`, e.g.

```
$ terraform import aws_api_gateway_documentation_version.example 5i4e1ko720/example-version
```
