---
subcategory: "ACM PCA"
layout: "aws"
page_title: "AWS: aws_acmpca_certificate_authority"
description: |-
  Provides a resource to manage AWS Certificate Manager Private Certificate Authorities
---

# Resource: aws_acmpca_certificate_authority

Provides a resource to manage AWS Certificate Manager Private Certificate Authorities (ACM PCA Certificate Authorities).

~> **NOTE:** Creating this resource will leave the certificate authority in a `PENDING_CERTIFICATE` status, which means it cannot yet issue certificates. To complete this setup, you must fully sign the certificate authority CSR available in the `certificate_signing_request` attribute and import the signed certificate outside of Terraform. Terraform can support another resource to manage that workflow automatically in the future.

## Example Usage

### Basic

```hcl
resource "aws_acmpca_certificate_authority" "example" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  permanent_deletion_time_in_days = 7
}
```

### Enable Certificate Revocation List

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

data "aws_iam_policy_document" "acmpca_bucket_access" {
  statement {
    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      aws_s3_bucket.example.arn,
      "${aws_s3_bucket.example.arn}/*",
    ]

    principals {
      identifiers = ["acm-pca.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = data.aws_iam_policy_document.acmpca_bucket_access.json
}

resource "aws_acmpca_certificate_authority" "example" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }

  revocation_configuration {
    crl_configuration {
      custom_cname       = "crl.example.com"
      enabled            = true
      expiration_in_days = 7
      s3_bucket_name     = aws_s3_bucket.example.id
    }
  }

  depends_on = [aws_s3_bucket_policy.example]
}
```

## Argument Reference

The following arguments are supported:

* `certificate_authority_configuration` - (Required) Nested argument containing algorithms and certificate subject information. Defined below.
* `enabled` - (Optional) Whether the certificate authority is enabled or disabled. Defaults to `true`.
* `revocation_configuration` - (Optional) Nested argument containing revocation configuration. Defined below.
* `tags` - (Optional) Specifies a key-value map of user-defined tags that are attached to the certificate authority.
* `type` - (Optional) The type of the certificate authority. Defaults to `SUBORDINATE`. Valid values: `ROOT` and `SUBORDINATE`.
* `permanent_deletion_time_in_days` - (Optional) The number of days to make a CA restorable after it has been deleted, must be between 7 to 30 days, with default to 30 days.

### certificate_authority_configuration

* `key_algorithm` - (Required) Type of the public key algorithm and size, in bits, of the key pair that your key pair creates when it issues a certificate. Valid values can be found in the [ACM PCA Documentation](https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_CertificateAuthorityConfiguration.html).
* `signing_algorithm` - (Required) Name of the algorithm your private CA uses to sign certificate requests. Valid values can be found in the [ACM PCA Documentation](https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_CertificateAuthorityConfiguration.html).
* `subject` - (Required) Nested argument that contains X.500 distinguished name information. At least one nested attribute must be specified.

#### subject

Contains information about the certificate subject. Identifies the entity that owns or controls the public key in the certificate. The entity can be a user, computer, device, or service.

* `common_name` - (Optional) Fully qualified domain name (FQDN) associated with the certificate subject. Must be less than or equal to 64 characters in length.
* `country` - (Optional) Two digit code that specifies the country in which the certificate subject located. Must be less than or equal to 2 characters in length.
* `distinguished_name_qualifier` - (Optional) Disambiguating information for the certificate subject. Must be less than or equal to 64 characters in length.
* `generation_qualifier` - (Optional) Typically a qualifier appended to the name of an individual. Examples include Jr. for junior, Sr. for senior, and III for third. Must be less than or equal to 3 characters in length.
* `given_name` - (Optional) First name. Must be less than or equal to 16 characters in length.
* `initials` - (Optional) Concatenation that typically contains the first letter of the `given_name`, the first letter of the middle name if one exists, and the first letter of the `surname`. Must be less than or equal to 5 characters in length.
* `locality` - (Optional) The locality (such as a city or town) in which the certificate subject is located. Must be less than or equal to 128 characters in length.
* `organization` - (Optional) Legal name of the organization with which the certificate subject is affiliated. Must be less than or equal to 64 characters in length.
* `organizational_unit` - (Optional) A subdivision or unit of the organization (such as sales or finance) with which the certificate subject is affiliated. Must be less than or equal to 64 characters in length.
* `pseudonym` - (Optional) Typically a shortened version of a longer `given_name`. For example, Jonathan is often shortened to John. Elizabeth is often shortened to Beth, Liz, or Eliza. Must be less than or equal to 128 characters in length.
* `state` - (Optional) State in which the subject of the certificate is located. Must be less than or equal to 128 characters in length.
* `surname` - (Optional) Family name. In the US and the UK for example, the surname of an individual is ordered last. In Asian cultures the surname is typically ordered first. Must be less than or equal to 40 characters in length.
* `title` - (Optional) A title such as Mr. or Ms. which is pre-pended to the name to refer formally to the certificate subject. Must be less than or equal to 64 characters in length.

### revocation_configuration

* `crl_configuration` - (Optional) Nested argument containing configuration of the certificate revocation list (CRL), if any, maintained by the certificate authority. Defined below.

#### crl_configuration

* `custom_cname` - (Optional) Name inserted into the certificate CRL Distribution Points extension that enables the use of an alias for the CRL distribution point. Use this value if you don't want the name of your S3 bucket to be public. Must be less than or equal to 253 characters in length.
* `enabled` - (Optional) Boolean value that specifies whether certificate revocation lists (CRLs) are enabled. Defaults to `false`.
* `expiration_in_days` - (Required) Number of days until a certificate expires. Must be between 1 and 5000.
* `s3_bucket_name` - (Optional) Name of the S3 bucket that contains the CRL. If you do not provide a value for the `custom_cname` argument, the name of your S3 bucket is placed into the CRL Distribution Points extension of the issued certificate. You must specify a bucket policy that allows ACM PCA to write the CRL to your bucket. Must be less than or equal to 255 characters in length.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the certificate authority.
* `arn` - Amazon Resource Name (ARN) of the certificate authority.
* `certificate` - Base64-encoded certificate authority (CA) certificate. Only available after the certificate authority certificate has been imported.
* `certificate_chain` - Base64-encoded certificate chain that includes any intermediate certificates and chains up to root on-premises certificate that you used to sign your private CA certificate. The chain does not include your private CA certificate. Only available after the certificate authority certificate has been imported.
* `certificate_signing_request` - The base64 PEM-encoded certificate signing request (CSR) for your private CA certificate.
* `not_after` - Date and time after which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `not_before` - Date and time before which the certificate authority is not valid. Only available after the certificate authority certificate has been imported.
* `serial` - Serial number of the certificate authority. Only available after the certificate authority certificate has been imported.
* `status` - Status of the certificate authority.

## Timeouts

`aws_acmpca_certificate_authority` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `1m`) How long to wait for a certificate authority to be created.

## Import

`aws_acmpca_certificate_authority` can be imported by using the certificate authority Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_acmpca_certificate_authority.example arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/12345678-1234-1234-1234-123456789012
```
