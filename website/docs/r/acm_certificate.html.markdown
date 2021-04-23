---
subcategory: "ACM"
layout: "aws"
page_title: "AWS: aws_acm_certificate"
description: |-
  Requests and manages a certificate from Amazon Certificate Manager (ACM).
---

# Resource: aws_acm_certificate

The ACM certificate resource allows requesting and management of certificates
from the Amazon Certificate Manager.

It deals with requesting certificates and managing their attributes and life-cycle.
This resource does not deal with validation of a certificate but can provide inputs
for other resources implementing the validation. It does not wait for a certificate to be issued.
Use a [`aws_acm_certificate_validation`](acm_certificate_validation.html) resource for this.

Most commonly, this resource is used together with [`aws_route53_record`](route53_record.html) and
[`aws_acm_certificate_validation`](acm_certificate_validation.html) to request a DNS validated certificate,
deploy the required validation records and wait for validation to complete.

Domain validation through email is also supported but should be avoided as it requires a manual step outside
of Terraform.

It's recommended to specify `create_before_destroy = true` in a [lifecycle][1] block to replace a certificate
which is currently in use (eg, by [`aws_lb_listener`](lb_listener.html)).

## Example Usage

### Certificate creation

```hcl
resource "aws_acm_certificate" "cert" {
  domain_name       = "example.com"
  validation_method = "DNS"

  tags = {
    Environment = "test"
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

### Importing an existing certificate

```hcl
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = tls_private_key.example.private_key_pem
  certificate_body = tls_self_signed_cert.example.cert_pem
}
```

### Referencing domain_validation_options With for_each Based Resources

See the [`aws_acm_certificate_validation` resource](acm_certificate_validation.html) for a full example of performing DNS validation.

```hcl
resource "aws_route53_record" "example" {
  for_each = {
    for dvo in aws_acm_certificate.example.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.example.zone_id
}
```

## Argument Reference

The following arguments are supported:

* Creating an amazon issued certificate
    * `domain_name` - (Required) A domain name for which the certificate should be issued
    * `subject_alternative_names` - (Optional) Set of domains that should be SANs in the issued certificate. To remove all elements of a previously configured list, set this value equal to an empty list (`[]`) or use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html) to trigger recreation.
    * `validation_method` - (Required) Which method to use for validation. `DNS` or `EMAIL` are valid, `NONE` can be used for certificates that were imported into ACM and then into Terraform.
    * `options` - (Optional) Configuration block used to set certificate options. Detailed below.
* Importing an existing certificate
    * `private_key` - (Required) The certificate's PEM-formatted private key
    * `certificate_body` - (Required) The certificate's PEM-formatted public key
    * `certificate_chain` - (Optional) The certificate's PEM-formatted chain
* Creating a private CA issued certificate
    * `domain_name` - (Required) A domain name for which the certificate should be issued
    * `certificate_authority_arn` - (Required) ARN of an ACMPCA
    * `subject_alternative_names` - (Optional) Set of domains that should be SANs in the issued certificate. To remove all elements of a previously configured list, set this value equal to an empty list (`[]`) or use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html) to trigger recreation.
* `tags` - (Optional) A map of tags to assign to the resource.

## options Configuration Block

Supported nested arguments for the `options` configuration block:

* `certificate_transparency_logging_preference` - (Optional) Specifies whether certificate details should be added to a certificate transparency log. Valid values are `ENABLED` or `DISABLED`. See https://docs.aws.amazon.com/acm/latest/userguide/acm-concepts.html#concept-transparency for more details.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the certificate
* `arn` - The ARN of the certificate
* `domain_name` - The domain name for which the certificate is issued
* `domain_validation_options` - Set of domain validation objects which can be used to complete certificate validation. Can have more than one element, e.g. if SANs are defined. Only set if `DNS`-validation was used.
* `status` - Status of the certificate.
* `validation_emails` - A list of addresses that received a validation E-Mail. Only set if `EMAIL`-validation was used.

Domain validation objects export the following attributes:

* `domain_name` - The domain to be validated
* `resource_record_name` - The name of the DNS record to create to validate the certificate
* `resource_record_type` - The type of DNS record to create
* `resource_record_value` - The value the DNS record needs to have

[1]: https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html

## Import

Certificates can be imported using their ARN, e.g.

```
$ terraform import aws_acm_certificate.cert arn:aws:acm:eu-central-1:123456789012:certificate/7e7a28d2-163f-4b8f-b9cd-822f96c08d6a
```
