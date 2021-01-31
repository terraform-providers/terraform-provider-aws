---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_domain_mail_from"
description: |-
  Provides an SES domain MAIL FROM resource
---

# Resource: aws_ses_domain_mail_from

Provides an SES domain MAIL FROM resource.

~> **NOTE:** For the MAIL FROM domain to be fully usable, this resource should be paired with the [aws_ses_domain_identity resource](/docs/providers/aws/r/ses_domain_identity.html). To validate the MAIL FROM domain, a DNS MX record is required. To pass SPF checks, a DNS TXT record may also be required. See the [Amazon SES MAIL FROM documentation](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/mail-from-set.html) for more information.

## Example Usage

```hcl
resource "aws_ses_domain_mail_from" "example" {
  domain           = aws_ses_domain_identity.example.domain
  mail_from_domain = "bounce.${aws_ses_domain_identity.example.domain}"
}

# Example SES Domain Identity
resource "aws_ses_domain_identity" "example" {
  domain = "example.com"
}

# Example Route53 MX record
resource "aws_route53_record" "example_ses_domain_mail_from_mx" {
  zone_id = aws_route53_zone.example.id
  name    = aws_ses_domain_mail_from.example.mail_from_domain
  type    = "MX"
  ttl     = "600"
  records = ["10 feedback-smtp.us-east-1.amazonses.com"] # Change to the region in which `aws_ses_domain_identity.example` is created
}

# Example Route53 TXT record for SPF
resource "aws_route53_record" "example_ses_domain_mail_from_txt" {
  zone_id = aws_route53_zone.example.id
  name    = aws_ses_domain_mail_from.example.mail_from_domain
  type    = "TXT"
  ttl     = "600"
  records = ["v=spf1 include:amazonses.com -all"]
}
```

## Argument Reference

The following arguments are required:

* `domain` - (Required) Verified domain name to generate DKIM tokens for.
* `mail_from_domain` - (Required) Subdomain (of above domain) which is to be used as MAIL FROM address (Required for DMARC validation)

The following arguments are optional:

* `behavior_on_mx_failure` - (Optional) The action that you want Amazon SES to take if it cannot successfully read the required MX record when you send an email. Defaults to `UseDefaultValue`. See the [SES API documentation](https://docs.aws.amazon.com/ses/latest/APIReference/API_SetIdentityMailFromDomain.html) for more information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The domain name.

## Import

MAIL FROM domain can be imported using the `domain` attribute, e.g.

```
$ terraform import aws_ses_domain_mail_from.example example.com
```
