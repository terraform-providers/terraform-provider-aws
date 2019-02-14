---
layout: "aws"
page_title: "AWS: aws_iot_certificate"
sidebar_current: "docs-aws-resource-iot-certificate"
description: |-
    Creates and manages an AWS IoT certificate.
---

# aws_iot_certificate

Creates and manages an AWS IoT certificate.

## Example Usage

```hcl
resource "aws_iot_certificate" "cert" {
  csr    = "${file("/my/csr.pem")}"
  active = true
}
```

## Argument Reference

* `active` - (Required)  Boolean flag to indicate if the certificate should be active
* `csr` - (Required) The certificate signing request. Review the
  [IoT API Reference Guide] (http://docs.aws.amazon.com/iot/latest/apireference/API_CreateCertificateFromCsr.html)
  for more information on creating a certificate from a certificate signing request (CSR).


## Attributes Reference

* `arn` - The ARN of the created AWS IoT certificate
