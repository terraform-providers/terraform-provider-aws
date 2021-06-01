---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_domain_name"
description: |-
 Get information on a custom domain name for use with AWS API Gateway.
---

# Data Source: aws_api_gateway_domain_name

Use this data source to get the custom domain name for use with AWS API Gateway.

## Example Usage

```hcl
data "aws_api_gateway_domain_name" "example" {
  domain_name = "api.example.com"
}
```

## Argument Reference

* `domain_name` - (Required) The fully-qualified domain name to look up. If no domain name is found, an error will be returned.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `arn` - The ARN of the found custom domain name.
* `certificate_arn` - The ARN for an AWS-managed certificate that is used by edge-optimized endpoint for this domain name.
* `certificate_name` - The name of the certificate that is used by edge-optimized endpoint for this domain name.
* `certificate_upload_date` - The upload date associated with the domain certificate.
* `cloudfront_domain_name` - The hostname created by Cloudfront to represent the distribution that implements this domain name mapping.
* `cloudfront_zone_id` - For convenience, the hosted zone ID (`Z2FDTNDATAQYW2`) that can be used to create a Route53 alias record for the distribution.
* `endpoint_configuration` - List of objects with the endpoint configuration of this domain name.
    * `types` - List of endpoint types.
* `regional_certificate_arn` - The ARN for an AWS-managed certificate that is used for validating the regional domain name.
* `regional_certificate_name` - The user-friendly name of the certificate that is used by regional endpoint for this domain name.
* `regional_domain_name` - The hostname for the custom domain's regional endpoint.
* `regional_zone_id` - The hosted zone ID that can be used to create a Route53 alias record for the regional endpoint.
* `security_policy` - The security policy for the domain name.
* `tags` - Key-value map of tags for the resource.
