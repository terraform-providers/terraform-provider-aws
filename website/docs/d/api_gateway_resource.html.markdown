---
layout: "aws"
page_title: "AWS: aws_api_gateway_resource"
sidebar_current: "docs-aws_api_gateway_resource"
description: |-
  Get information on a API Gateway Resource
---

# Data Source: aws_api_gateway_resource

Use this data source to get the id of a Resource in API Gateway. 
To fetch the Resource, you must provide the REST API id as well as the full path.  

## Example Usage

```hcl
data "aws_api_gateway_rest_api" "my_rest_api" {
  name = "my-rest-api"
}

data "aws_api_gateway_resource" "my_resource" {
  rest_api_id = "${aws_api_gateway_rest_api.my_rest_api.id}"
  path        = "/endpoint/path"
}
```

## Argument Reference

 * `rest_api_id` - (Required) The REST API id that owns the resource. If no REST API is found, an error will be returned.
 * `path` - (Required) The full path of the resource.  If no path is found, an error will be returned.

## Attributes Reference

 * `id` - Set to the ID of the found Resource.
 * `parent_id` - Set to the ID of the parent Resource.
 * `path_part` - Set to the path relative to the parent Resource.
