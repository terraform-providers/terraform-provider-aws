---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_registration_code"
description: |-
  Get the unique IoT registration code
---

# Data Source: aws_iot_registration_code

Returns a unique registration code specific to the AWS account making the call.

## Example Usage

```hcl
data "aws_iot_registration_code" "example" {}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  key_algorithm   = "RSA"
  private_key_pem = tls_private_key.verification.private_key_pem

  subject {
    common_name = data.aws_iot_registration_code.example.id
  }
}
```

## Attributes Reference

* `code` - The code required to be set in a verification common name.
