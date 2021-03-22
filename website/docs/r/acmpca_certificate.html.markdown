---
subcategory: "ACM PCA"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate"
description: |-
  Provides a resource to issue a certificate using AWS Certificate Manager Private Certificate Authority (ACM PCA)
---

# Resource: aws_acmpca_certificate

Provides a resource to issue a certificate using AWS Certificate Manager Private Certificate Authority (ACM PCA).

## Example Usage

### Basic

```terraform
resource "aws_acmpca_certificate" "example" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.example.arn
  certificate_signing_request = tls_cert_request.csr.cert_request_pem
  signing_algorithm           = "SHA256WITHRSA"
  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "example" {
  private_certificate_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  permanent_deletion_time_in_days = 7
}

resource "tls_private_key" "key" {
  algorithm = "RSA"
}

resource "tls_cert_request" "csr" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.key.private_key_pem

  subject {
    common_name = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `certificate_authority_arn` - (Required) Amazon Resource Name (ARN) of the certificate authority.
* `certificate_signing_request` - (Required) Certificate Signing Request in PEM format.
* `signing_algorithm` - (Required) Algorithm to use to sign certificate requests. Valid values: `SHA256WITHRSA`, `SHA256WITHECDSA`, `SHA384WITHRSA`, `SHA384WITHECDSA`, `SHA512WITHRSA`, `SHA512WITHECDSA`
* `validity` - (Required) Configures end of the validity period for the certificate. See [validity block](#validity-block) below.
* `template_arn` - (Optional) The template to use when issuing a certificate. See [ACM PCA Documentation](https://docs.aws.amazon.com/acm-pca/latest/userguide/UsingTemplates.html) for more information.

### validity block

* `type` - (Required) Determines how `value` is interpreted. Valid values: `DAYS`, `MONTHS`, `YEARS`, `ABSOLUTE`, `END_DATE`.
* `value` - (Required) If `type` is `DAYS`, `MONTHS`, or `YEARS`, the relative time until the certificate expires. If `type` is `ABSOLUTE`, the date in seconds since the Unix epoch. If `type` is `END_DATE`, the  date in RFC 3339 format.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the certificate.
* `certificate` - The PEM-encoded certificate value.
* `certificate_chain` - The PEM-encoded certificate chain that includes any intermediate certificates and chains up to root CA.

## Import

`aws_acmpca_certificate` can not be imported at this time.
