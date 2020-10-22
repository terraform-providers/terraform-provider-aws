---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_function"
description: |-
  Provides an AppSync Function.
---

# Resource: aws_appsync_function

Provides an AppSync Function.

## Example Usage

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
  schema              = <<EOF
type Mutation {
  putPost(id: ID!, title: String!): Post
}

type Post {
  id: ID!
  title: String!
}

type Query {
  singlePost(id: ID!): Post
}

schema {
  query: Query
  mutation: Mutation
}
EOF
}

resource "aws_appsync_datasource" "example" {
  api_id = aws_appsync_graphql_api.example.id
  name   = "example"
  type   = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}

resource "aws_appsync_function" "example" {
  api_id                   = aws_appsync_graphql_api.example.id
  data_source              = aws_appsync_datasource.example.name
  name                     = "example"
  request_mapping_template = <<EOF
{
    "version": "2018-05-29",
    "method": "GET",
    "resourcePath": "/",
    "params":{
        "headers": $utils.http.copyheaders($ctx.request.headers)
    }
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
    $ctx.result.body
#else
    $utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The ID of the associated AppSync API.
* `data_source` - (Required) The Function DataSource name.
* `name` - (Required) The Function name. The function name does not have to be unique.
* `request_mapping_template` - (Required) The Function request mapping template. Functions support only the 2018-05-29 version of the request mapping template.
* `response_mapping_template` - (Required) The Function response mapping template.
* `description` - (Optional) The Function description.
* `function_version` - (Optional) The version of the request mapping template. Currently the supported value is `2018-05-29`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - API Function ID (Formatted as ApiId-FunctionId)
* `arn` - The ARN of the Function object.
* `function_id` - A unique ID representing the Function object.

## Import

`aws_appsync_function` can be imported using the AppSync API ID and Function ID separated by `-`, e.g.

```
$ terraform import aws_appsync_function.example xxxxx-yyyyy
```
