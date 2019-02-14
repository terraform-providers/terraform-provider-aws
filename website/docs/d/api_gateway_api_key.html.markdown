---
layout: "aws"
page_title: "AWS: aws_api_gateway_api_key"
sidebar_current: "docs-aws_api_gateway_api_key"
description: |-
  Get information on an API Gateway API Key
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
