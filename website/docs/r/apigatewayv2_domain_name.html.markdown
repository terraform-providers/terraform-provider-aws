---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_domain_name"
description: |-
  Manages an Amazon API Gateway Version 2 domain name.
---

# Resource: aws_apigatewayv2_domain_name

Manages an Amazon API Gateway Version 2 domain name.
More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html).

-> **Note:** This resource establishes ownership of and the TLS settings for
a particular domain name. An API stage can be associated with the domain name using the `aws_apigatewayv2_api_mapping` resource.

## Example Usage

### Basic

```hcl
resource "aws_apigatewayv2_domain_name" "example" {
  domain_name = "ws-api.example.com"

  domain_name_configuration {
    certificate_arn = "${aws_acm_certificate.example.arn}"
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The domain name.
* `domain_name_configuration` - (Required) The domain name configuration.
* `tags` - (Optional) A map of tags to assign to the domain name.

The `domain_name_configuration` object supports the following:

* `certificate_arn` - (Required) The ARN of an AWS-managed certificate that will be used by the endpoint for the domain name. AWS Certificate Manager is the only supported source.
Use the [`aws_acm_certificate`](/docs/providers/aws/r/acm_certificate.html) resource to configure an ACM certificate.
* `endpoint_type` - (Required) The endpoint type. Valid values: `REGIONAL`.
* `security_policy` - (Required) The Transport Layer Security (TLS) version of the [security policy](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-custom-domain-tls-version.html) for the domain name. Valid values: `TLS_1_2`.
* `hosted_zone_id` - (Computed) The Amazon Route 53 Hosted Zone ID of the endpoint.
* `target_domain_name` - (Computed) The target domain name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The domain name identifier.
* `api_mapping_selection_expression` - The [API mapping selection expression](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-websocket-api-selection-expressions.html#apigateway-websocket-api-mapping-selection-expressions) for the domain name.
* `arn` - The ARN of the domain name.

## Timeouts

`aws_apigatewayv2_domain_name` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `update` - (Default `60 minutes`) Used for updating the domain name

## Import

`aws_apigatewayv2_domain_name` can be imported by using the domain name, e.g.

```
$ terraform import aws_apigatewayv2_domain_name.example ws-api.example.com
```
