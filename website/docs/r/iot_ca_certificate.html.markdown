---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_certificate"
description: |-
    Creates and manages an AWS IoT certificate.
---

# Resource: aws_iot_certificate

Creates and manages an AWS IoT certificate.

## Example Usage

```hcl
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem
  key_algorithm   = "RSA"

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
  is_ca_certificate = true
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.verification.private_key_pem

  subject {
    common_name = data.aws_iot_registration_code.example.id
  }
}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_locally_signed_cert" "verification" {
  cert_request_pem   = tls_cert_request.verification.cert_request_pem
  ca_private_key_pem = tls_private_key.ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca.cert_pem
  ca_key_algorithm   = "RSA"

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_iot_ca_certificate" "example" {
  active                       = true
  ca_certificate_pem           = tls_self_signed_cert.ca.cert_pem
  verification_certificate_pem = tls_locally_signed_cert.verification.cert_pem
  auto_registration            = true
}

data "aws_iot_registration_code" "example" {}
```


## Argument Reference

* `active` - (Required)  Boolean flag to indicate if the certificate should be active for device authentication.
* `auto_registration` - (Required)  Boolean flag to indicate if the certificate should be active for device regisration.
* `ca_certificate_pem` - (Required)  PEM encoded ca certificate.
* `verification_certificate_pem` - (Required)  PEM encoded verification certificate containing the common name of a registration code. Review
  [CreateVerificationCSR](https://docs.aws.amazon.com/iot/latest/developerguide/register-CA-cert.html)

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `id` - The internal ID assigned to this ca certificate.
* `arn` - The ARN of the created ca certificate.
