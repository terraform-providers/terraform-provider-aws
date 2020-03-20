---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_email_identity"
description: |-
  Provides an SES email identity resource
---

# Resource: aws_ses_email_identity

Provides an SES email identity resource

## Argument Reference

The following arguments are supported:

* `email` - (Required) The email address to assign to SES

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the email identity.

## Example Usage

```hcl
resource "aws_ses_email_identity" "example" {
  email = "email@example.com"
}
```

## Import

SES email identities can be imported using the email address.

```
$ terraform import aws_ses_email_identity.example email@example.com
```
