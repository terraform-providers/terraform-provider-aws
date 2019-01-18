---
layout: "aws"
page_title: "AWS: aws_acm_certificate"
sidebar_current: "docs-aws-resource-acm-certificate"
description: |-
  Requests and manages a certificate from Amazon Certificate Manager (ACM).
---

# aws_acm_certificate

The ACM certificate resource allows requesting and management of certificates
from the Amazon Certificate Manager.

It deals with requesting certificates and managing their attributes and life-cycle.
This resource does not deal with validation of a certificate but can provide inputs
for other resources implementing the validation. It does not wait for a certificate to be issued.
Use a [`aws_acm_certificate_validation`](acm_certificate_validation.html) resource for this.

Most commonly, this resource is used to together with [`aws_route53_record`](route53_record.html) and
[`aws_acm_certificate_validation`](acm_certificate_validation.html) to request a DNS validated certificate,
deploy the required validation records and wait for validation to complete.

Domain validation through E-Mail is also supported but should be avoided as it requires a manual step outside
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

### Importation of existing certificate

```hcl
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

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
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

```

## Argument Reference

The following arguments are supported:

* Creating an amazon issued certificate
  * `domain_name` - (Required) A domain name for which the certificate should be issued
  * `subject_alternative_names` - (Optional) A list of domains that should be SANs in the issued certificate
  * `validation_method` - (Required) Which method to use for validation. `DNS` or `EMAIL` are valid, `NONE` can be used for certificates that were imported into ACM and then into Terraform.
* Importing an existing certificate
  * `private_key` - (Required) The certificate's PEM-formatted private key
  * `certificate_body` - (Required) The certificate's PEM-formatted public key
  * `certificate_chain` - (Optional) The certificate's PEM-formatted chain
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the certificate
* `arn` - The ARN of the certificate
* `domain_name` - The domain name for which the certificate is issued
* `domain_validation_options` - A list of attributes to feed into other resources to complete certificate validation. Can have more than one element, e.g. if SANs are defined. Only set if `DNS`-validation was used.
* `validation_emails` - A list of addresses that received a validation E-Mail. Only set if `EMAIL`-validation was used.

Domain validation objects export the following attributes:

* `domain_name` - The domain to be validated
* `resource_record_name` - The name of the DNS record to create to validate the certificate
* `resource_record_type` - The type of DNS record to create
* `resource_record_value` - The value the DNS record needs to have

[1]: /docs/configuration/resources.html#lifecycle

## Import

Certificates can be imported using their ARN, e.g.

```
$ terraform import aws_acm_certificate.cert arn:aws:acm:eu-central-1:123456789012:certificate/7e7a28d2-163f-4b8f-b9cd-822f96c08d6a
```