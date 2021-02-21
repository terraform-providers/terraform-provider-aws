---
subcategory: "ACM"
layout: "aws"
page_title: "AWS: aws_acm_certificate"
description: |-
  Get information on a Amazon Certificate Manager (ACM) Certificate
---

# Data Source: aws_acm_certificate

Use this data source to get the ARN of a certificate in AWS Certificate
Manager (ACM), you can reference
it by domain without having to hard code the ARNs as input.

## Example Usage

```hcl
# Find a certificate that is issued
data "aws_acm_certificate" "issued" {
  domain   = "tf.example.com"
  statuses = ["ISSUED"]
}

# Find a certificate issued by (not imported into) ACM
data "aws_acm_certificate" "amazon_issued" {
  domain      = "tf.example.com"
  types       = ["AMAZON_ISSUED"]
  most_recent = true
}

# Find a RSA 4096 bit certificate
data "aws_acm_certificate" "rsa_4096" {
  domain    = "tf.example.com"
  key_types = ["RSA_4096"]
}
```

## Argument Reference

* `domain` - (Required) The domain of the certificate to look up. If no certificate is found with this name, an error will be returned.
* `key_types` - (Optional) A list of key algorithms to filter certificates. By default, ACM does not return all certificate types when searching. Valid values are `RSA_1024`, `RSA_2048`, `RSA_4096`, `EC_prime256v1`, `EC_secp384r1`, and `EC_secp521r1`.
* `statuses` - (Optional) A list of statuses on which to filter the returned list. Valid values are `PENDING_VALIDATION`, `ISSUED`,
   `INACTIVE`, `EXPIRED`, `VALIDATION_TIMED_OUT`, `REVOKED` and `FAILED`. If no value is specified, only certificates in the `ISSUED` state
   are returned.
* `types` - (Optional) A list of types on which to filter the returned list. Valid values are `AMAZON_ISSUED` and `IMPORTED`.
* `most_recent` - (Optional) If set to true, it sorts the certificates matched by previous criteria by the NotBefore field, returning only the most recent one. If set to false, it returns an error if more than one certificate is found. Defaults to false.

## Attributes Reference

* `arn` - Amazon Resource Name (ARN) of the found certificate, suitable for referencing in other resources that support ACM certificates.
* `id` - Amazon Resource Name (ARN) of the found certificate, suitable for referencing in other resources that support ACM certificates.
* `tags` - A mapping of tags for the resource.

