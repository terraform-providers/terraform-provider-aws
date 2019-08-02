---
layout: "aws"
page_title: "AWS: aws_acm_certificate"
sidebar_current: "docs-aws-resource-acm-certificate"
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

#example with subject_alternative_names and domain_validation_options
resource "aws_acm_certificate" "cert" {
  domain_name               = "yolo.example.io"
  validation_method         = "EMAIL"
  subject_alternative_names = ["app1.yolo.example.io", "yolo.example.io"]

  domain_validation_options = [
    {
      domain_name       = "yolo.example.io"
      validation_domain = "example.io"
    },
    {
      domain_name       = "app1.yolo.example.io"
      validation_domain = "example.io"
    },
  ]
}

#basic example
resource "aws_acm_certificate" "cert" {
  domain_name               = "yolo.example.io"
  validation_method         = "EMAIL"
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
  * `domain_validaton_options` - (Optional) Contains information about the initial validation of each domain name that occurs. This is an array of maps that contains information about which validation_domain to use for domains in the subject_alternative_names list.
* Importing an existing certificate
  * `private_key` - (Required) The certificate's PEM-formatted private key
  * `certificate_body` - (Required) The certificate's PEM-formatted public key
  * `certificate_chain` - (Optional) The certificate's PEM-formatted chain
* `tags` - (Optional) A mapping of tags to assign to the resource.

Domain Validation Options objects accept the following attributes

* `domain_name` - (Required) A fully qualified domain name (FQDN) in the certificate. For example, www.example.com or example.com .
* `validation_domain` - (Required) The domain name that ACM used to send domain validation emails

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the certificate
* `arn` - The ARN of the certificate
* `certificate_details` - A list of attributes to feed into other resources to complete certificate validation. Can have more than one element, e.g. if SANs are defined. 

Certificate Details objects export the following attributes:

* `domain_name` - A fully qualified domain name (FQDN) in the certificate. For example, www.example.com or example.com .
* `resource_record_name` - The name of the DNS record to create in your domain. This is supplied by ACM.
* `resource_record_type` - The type of DNS record. Currently this can be CNAME .
* `resource_record_value` - The value of the CNAME record to add to your DNS database. This is supplied by ACM. 
* `validation_domain` - The domain name that ACM used to send domain validation emails.
* `validation_method` - One of EMAIl or DNS
* `validation_emails` - A list of email addresses that ACM used to send domain validation emails.

[1]: /docs/configuration/resources.html#lifecycle

## Import

Certificates can be imported using their ARN, e.g.

```
$ terraform import aws_acm_certificate.cert arn:aws:acm:eu-central-1:123456789012:certificate/7e7a28d2-163f-4b8f-b9cd-822f96c08d6a
```