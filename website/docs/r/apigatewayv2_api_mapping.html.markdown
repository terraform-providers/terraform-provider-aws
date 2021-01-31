---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_api_mapping"
description: |-
  Manages an Amazon API Gateway Version 2 API mapping.
---

# Resource: aws_apigatewayv2_api_mapping

Manages an Amazon API Gateway Version 2 API mapping.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_api_mapping" "example" {
  api_id      = aws_apigatewayv2_api.example.id
  domain_name = aws_apigatewayv2_domain_name.example.id
  stage       = aws_apigatewayv2_stage.example.id
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API identifier.
* `domain_name` - (Required) The domain name. Use the [`aws_apigatewayv2_domain_name`](/docs/providers/aws/r/apigatewayv2_domain_name.html) resource to configure a domain name.
* `stage` - (Required) The API stage. Use the [`aws_apigatewayv2_stage`](/docs/providers/aws/r/apigatewayv2_stage.html) resource to configure an API stage.
* `api_mapping_key` - (Optional) The [API mapping key](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-mapping-template-reference.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The API mapping identifier.

## Import

`aws_apigatewayv2_api_mapping` can be imported by using the API mapping identifier and domain name, e.g.

```
$ terraform import aws_apigatewayv2_api_mapping.example 1122334/ws-api.example.com
```
