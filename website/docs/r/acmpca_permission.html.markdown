---
subcategory: "ACM PCA"
layout: "aws"
page_title: "AWS: aws_acmpca_permission"
description: |-
  Provides a resource to manage an AWS Certificate Manager Private Certificate Authorities Permission
---

# Resource: aws_acmpca_permission

Provides a resource to manage an AWS Certificate Manager Private Certificate Authorities Permission.
Currently, this is only required in order to allow the ACM service to automatically certificates issued by a PCA.

## Example Usage

```hcl
resource "aws_acmpca_permission" "example" {
  certificate_authority_arn = "${aws_acmpca_certificate_authority.example.arn}"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
  principal                 = "acm.amazonaws.com"
}

resource "aws_acmpca_certificate_authority" "example" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `certificate_authority_arn` - (Required) The Amazon Resource Name (ARN) of the CA that grants the permissions.
* `actions` - (Required) The actions that the specified AWS service principal can use. These include `IssueCertificate`, `GetCertificate`, and `ListPermissions`. Note that in order for ACM to automatically rotate certificates issued by a PCA, it must be granted permission on all 3 actions, as per the example above.
* `principal` - (Required) The AWS service or identity that receives the permission. At this time, the only valid principal is `acm.amazonaws.com`.
* `source_account` - (Optional) The ID of the calling account
