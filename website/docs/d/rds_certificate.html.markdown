---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_certificate"
description: |-
  Information about an RDS Certificate.
---

# Data Source: aws_rds_certificate

Information about an RDS Certificate.

## Example Usage

```hcl
data "aws_rds_certificate" "example" {
  latest_valid_till = true
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) Certificate identifier. For example, `rds-ca-2019`.
* `latest_valid_till` - (Optional) When enabled, returns the certificate with the latest `ValidTill`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the certificate.
* `certificate_type` - Type of certificate. For example, `CA`.
* `customer_override` - Boolean whether there is an override for the default certificate identifier.
* `customer_override_valid_till` - If there is an override for the default certificate identifier, when the override expires.
* `thumbprint` - Thumbprint of the certificate.
* `valid_from` - [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of certificate starting validity date.
* `valid_till` - [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of certificate ending validity date.
