---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_identity_feedback_forwarding_enabled"
description: |-
  Enable and Disable identity feedback forwarding
---

# Resource: aws_ses_identity_feedback_forwarding_enabled

Enable and Disable identity feedback forwarding

## Example Usage

```hcl
resource "aws_ses_domain_identity" "example" {
  domain = "example.com"
}

resource "aws_ses_identity_feedback_forwarding_enabled" "example" {
  identity = aws_ses_domain_identity.example.domain
  enabled  = false
}
```

## Argument Reference

The following arguments are supported:

* `identity` - (Required) Domain name or E-mail address which have already been registered in SES
* `enabled` - (Required) Whether you enable feedback forwarding or not
