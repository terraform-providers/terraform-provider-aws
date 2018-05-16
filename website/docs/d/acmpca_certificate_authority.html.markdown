---
layout: "aws"
page_title: "AWS: aws_acmpca_certificate_authority"
sidebar_current: "docs-aws-datasource-acmpca-certificate-authority"
description: |-
  Get information on a AWS Certificate Manager Private Certificate Authority
---

# Data Source: aws_acmpca_certificate_authority

Get information on a AWS Certificate Manager Private Certificate Authority (ACM PCA Certificate Authority).

## Example Usage

```hcl
data "aws_acmpca_certificate_authority" "example" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) Amazon Resource Name (ARN) of the certificate authority.

## Attribute Reference

The following additional attributes are exported:

* `id` - Amazon Resource Name (ARN) of the certificate authority.
* `certificate` - Base64-encoded certificate authority (CA) certificate. Only available after the certificate authority certificate has been imported.
* `certificate_chain` - Base64-encoded certificate chain that includes any intermediate certificates and chains up to root on-premises certificate that you used to sign your private CA certificate. The chain does not include your private CA certificate. Only available after the certificate authority certificate has been imported.
* `certificate_signing_request` - The base64 PEM-encoded certificate signing request (CSR) for your private CA certificate.
* `not_after` - Date and time after which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `not_before` - Date and time before which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `revocation_configuration` - Nested attribute containing revocation configuration.
  * `revocation_configuration.0.crl_configuration` - Nested attribute containing configuration of the certificate revocation list (CRL), if any, maintained by the certificate authority.
    * `revocation_configuration.0.crl_configuration.0.custom_cname` - Name inserted into the certificate CRL Distribution Points extension that enables the use of an alias for the CRL distribution point.
    * `revocation_configuration.0.crl_configuration.0.enabled` - Boolean value that specifies whether certificate revocation lists (CRLs) are enabled.
    * `revocation_configuration.0.crl_configuration.0.expiration_in_days` - Number of days until a certificate expires.
    * `revocation_configuration.0.crl_configuration.0.s3_bucket_name` - Name of the S3 bucket that contains the CRL.
* `serial` - Serial number of the certificate authority. Only available after the certificate authority certificate has been imported.
* `status` - Status of the certificate authority.
* `tags` - Specifies a key-value map of user-defined tags that are attached to the certificate authority.
* `type` - The type of the certificate authority.
