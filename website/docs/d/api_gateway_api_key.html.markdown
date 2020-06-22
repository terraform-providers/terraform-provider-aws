---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_api_key"
description: |-
  Get information on an API Gateway REST API Key
---

# Data Source: aws_api_gateway_api_key

Use this data source to get the name and value of a pre-existing API Key, for
example to supply credentials for a dependency microservice.

## Example Usage

```hcl
data "aws_api_gateway_api_key" "my_api_key" {
  id = "ru3mpjgse6"
}
```

## Argument Reference

 * `id` - (Required) The ID of the API Key to look up.

## Attributes Reference

 * `id` - Set to the ID of the API Key.
 * `name` - Set to the name of the API Key.
 * `value` - Set to the value of the API Key.
 * `created_date` - The date and time when the API Key was created.
 * `last_updated_date` - The date and time when the API Key was last updated.
 * `description` - The description of the API Key.
 * `enabled` - Specifies whether the API Key is enabled.
 * `tags` - A map of tags for the resource.
