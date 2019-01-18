---
layout: "aws"
page_title: "AWS: aws_api_gateway_domain_name"
sidebar_current: "docs-aws-resource-api-gateway-domain-name"
description: |-
  Registers a custom domain name for use with AWS API Gateway.
---

# aws_api_gateway_domain_name

Registers a custom domain name for use with AWS API Gateway.

This resource just establishes ownership of and the TLS settings for
a particular domain name. An API can be attached to a particular path
under the registered domain name using
[the `aws_api_gateway_base_path_mapping` resource](api_gateway_base_path_mapping.html).

API Gateway domains can be defined as either 'edge-optimized' or 'regional'.  In an edge-optimized configuration,
API Gateway internally creates and manages a CloudFront distribution to route requests on the given hostname. In
addition to this resource it's necessary to create a DNS record corresponding to the given domain name which is an alias
(either Route53 alias or traditional CNAME) to the Cloudfront domain name exported in the `cloudfront_domain_name`
attribute.

In a regional configuration, API Gateway does not create a CloudFront distribution to route requests to the API, though
a distribution can be created if needed. In either case, it is necessary to create a DNS record corresponding to the
given domain name which is an alias (either Route53 alias or traditional CNAME) to the regional domain name exported in
the `regional_domain_name` attribute.

~> **Note:** All arguments including the private key will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

-> For information about regions that support AWS Certificate Manager (ACM), see the [Regions and Endpoints Documentation](https://docs.aws.amazon.com/general/latest/gr/rande.html#acm_region).

### Edge Optimized (ACM Certificate)

```hcl
resource "aws_api_gateway_domain_name" "example" {
  certificate_arn = "${aws_acm_certificate_validation.example.certificate_arn}"
  domain_name     = "api.example.com"
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = "${aws_api_gateway_domain_name.example.domain_name}"
  type    = "A"
  zone_id = "${aws_route53_zone.example.id}"

  alias {
    evaluate_target_health = true
    name                   = "${aws_api_gateway_domain_name.example.cloudfront_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.example.cloudfront_zone_id}"
  }
}
```

### Edge Optimized (Uploaded Certificate)

```hcl
resource "aws_api_gateway_domain_name" "example" {
  domain_name = "api.example.com"

  certificate_name        = "example-api"
  certificate_body        = "${file("${path.module}/example.com/example.crt")}"
  certificate_chain       = "${file("${path.module}/example.com/ca.crt")}"
  certificate_private_key = "${file("${path.module}/example.com/example.key")}"
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  zone_id = "${aws_route53_zone.example.id}" # See aws_route53_zone for how to create this

  name = "${aws_api_gateway_domain_name.example.domain_name}"
  type = "A"

  alias {
    name                   = "${aws_api_gateway_domain_name.example.cloudfront_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.example.cloudfront_zone_id}"
    evaluate_target_health = true
  }
}
```

### Regional (ACM Certificate)

```hcl
resource "aws_api_gateway_domain_name" "example" {
  domain_name              = "api.example.com"
  regional_certificate_arn = "${aws_acm_certificate_validation.example.certificate_arn}"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = "${aws_api_gateway_domain_name.example.domain_name}"
  type    = "A"
  zone_id = "${aws_route53_zone.example.id}"

  alias {
    evaluate_target_health = true
    name                   = "${aws_api_gateway_domain_name.example.regional_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.example.regional_zone_id}"
  }
}
```

### Regional (Uploaded Certificate)

```hcl
resource "aws_api_gateway_domain_name" "example" {
  certificate_body          = "${file("${path.module}/example.com/example.crt")}"
  certificate_chain         = "${file("${path.module}/example.com/ca.crt")}"
  certificate_private_key   = "${file("${path.module}/example.com/example.key")}"
  domain_name               = "api.example.com"
  regional_certificate_name = "example-api"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# Example DNS record using Route53.
# Route53 is not specifically required; any DNS host can be used.
resource "aws_route53_record" "example" {
  name    = "${aws_api_gateway_domain_name.example.domain_name}"
  type    = "A"
  zone_id = "${aws_route53_zone.example.id}"

  alias {
    evaluate_target_health = true
    name                   = "${aws_api_gateway_domain_name.example.regional_domain_name}"
    zone_id                = "${aws_api_gateway_domain_name.example.regional_zone_id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The fully-qualified domain name to register
* `endpoint_configuration` - (Optional) Configuration block defining API endpoint information including type. Defined below.

When referencing an AWS-managed certificate, the following arguments are supported:

* `certificate_arn` - (Optional) The ARN for an AWS-managed certificate. AWS Certificate Manager is the only supported source. Used when an edge-optimized domain name is desired. Conflicts with `certificate_name`, `certificate_body`, `certificate_chain`, `certificate_private_key`, `regional_certificate_arn`, and `regional_certificate_name`.
* `regional_certificate_arn` - (Optional) The ARN for an AWS-managed certificate. AWS Certificate Manager is the only supported source. Used when a regional domain name is desired. Conflicts with `certificate_arn`, `certificate_name`, `certificate_body`, `certificate_chain`, and `certificate_private_key`.

When uploading a certificate, the following arguments are supported:

* `certificate_name` - (Optional) The unique name to use when registering this
  certificate as an IAM server certificate. Conflicts with `certificate_arn`, `regional_certificate_arn`, and
  `regional_certificate_name`. Required if `certificate_arn` is not set.
* `certificate_body` - (Optional) The certificate issued for the domain name
  being registered, in PEM format. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`, `regional_certificate_arn`, and
  `regional_certificate_name`.
* `certificate_chain` - (Optional) The certificate for the CA that issued the
  certificate, along with any intermediate CA certificates required to
  create an unbroken chain to a certificate trusted by the intended API clients. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`,
  `regional_certificate_arn`, and `regional_certificate_name`.
* `certificate_private_key` - (Optional) The private key associated with the
  domain certificate given in `certificate_body`. Only valid for `EDGE` endpoint configuration type. Conflicts with `certificate_arn`, `regional_certificate_arn`, and `regional_certificate_name`.
* `regional_certificate_name` - (Optional) The user-friendly name of the certificate that will be used by regional endpoint for this domain name. Conflicts with `certificate_arn`, `certificate_name`, `certificate_body`, `certificate_chain`, and
  `certificate_private_key`.

### endpoint_configuration

* `types` - (Required) A list of endpoint types. This resource currently only supports managing a single value. Valid values: `EDGE` or `REGIONAL`. If unspecified, defaults to `EDGE`. Must be declared as `REGIONAL` in non-Commercial partitions. Refer to the [documentation](https://docs.aws.amazon.com/apigateway/latest/developerguide/create-regional-api.html) for more information on the difference between edge-optimized and regional APIs.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `id` - The internal id assigned to this domain name by API Gateway.
* `certificate_upload_date` - The upload date associated with the domain certificate.
* `cloudfront_domain_name` - The hostname created by Cloudfront to represent
  the distribution that implements this domain name mapping.
* `cloudfront_zone_id` - For convenience, the hosted zone ID (`Z2FDTNDATAQYW2`)
  that can be used to create a Route53 alias record for the distribution.
* `regional_domain_name` - The hostname for the custom domain's regional endpoint.
* `regional_zone_id` - The hosted zone ID that can be used to create a Route53 alias record for the regional endpoint.

## Import

API Gateway domain names can be imported using their `name`, e.g.

```
$ terraform import aws_api_gateway_domain_name.example dev.example.com
```
