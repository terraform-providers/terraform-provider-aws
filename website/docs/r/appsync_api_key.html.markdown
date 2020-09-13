---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_api_key"
description: |-
  Provides an AppSync API Key.
---

# Resource: aws_appsync_api_key

Provides an AppSync API Key.

## Example Usage

```hcl
resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "example"
}

resource "aws_appsync_api_key" "example" {
  api_id  = aws_appsync_graphql_api.example.id
  expires = "2018-05-03T04:00:00Z"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The ID of the associated AppSync API
* `description` - (Optional) The API key description. Defaults to "Managed by Terraform".
* `expires` - (Optional) RFC3339 string representation of the expiry date. Rounded down to nearest hour. By default, it is 7 days from the date of creation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - API Key ID (Formatted as ApiId:Key)
* `key` - The API key

## Import

`aws_appsync_api_key` can be imported using the AppSync API ID and key separated by `:`, e.g.

```
$ terraform import aws_appsync_api_key.example xxxxx:yyyyy
```
