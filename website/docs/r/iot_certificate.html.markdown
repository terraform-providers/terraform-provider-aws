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

### With CSR

```hcl
resource "aws_iot_certificate" "cert" {
  csr    = file("/my/csr.pem")
  active = true
}
```

### Without CSR

```hcl
resource "aws_iot_certificate" "cert" {
  active = true
}
```

## Argument Reference

* `active` - (Required)  Boolean flag to indicate if the certificate should be active
* `csr` - (Optional) The certificate signing request. Review
  [CreateCertificateFromCsr](https://docs.aws.amazon.com/iot/latest/apireference/API_CreateCertificateFromCsr.html)
  for more information on generating a certificate from a certificate signing request (CSR).
  If none is specified both the certificate and keys will be generated, review [CreateKeysAndCertificate](https://docs.aws.amazon.com/iot/latest/apireference/API_CreateKeysAndCertificate.html)
  for more information on generating keys and a certificate.

## Attributes Reference

In addition to the arguments, the following attributes are exported:

* `id` - The internal ID assigned to this certificate.
* `arn` - The ARN of the created certificate.
* `certificate_pem` - The certificate data, in PEM format.
* `public_key` - When no CSR is provided, the public key.
* `private_key` - When no CSR is provided, the private key.

