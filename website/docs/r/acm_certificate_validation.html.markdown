---
subcategory: "ACM"
layout: "aws"
page_title: "AWS: aws_acm_certificate_validation"
description: |-
  Waits for and checks successful validation of an ACM certificate.
---

# Resource: aws_acm_certificate_validation

This resource represents a successful validation of an ACM certificate in concert
with other resources.

Most commonly, this resource is used together with [`aws_route53_record`](route53_record.html) and
[`aws_acm_certificate`](acm_certificate.html) to request a DNS validated certificate,
deploy the required validation records and wait for validation to complete.

~> **WARNING:** This resource implements a part of the validation workflow. It does not represent a real-world entity in AWS, therefore changing or deleting this resource on its own has no immediate effect.


## Example Usage

### DNS Validation with Route 53

```hcl
resource "aws_acm_certificate" "example" {
  domain_name       = "example.com"
  validation_method = "DNS"
}

data "aws_route53_zone" "example" {
  name         = "example.com"
  private_zone = false
}

resource "aws_route53_record" "example" {
  for_each = {
    for dvo in aws_acm_certificate.example.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.example.zone_id
}

resource "aws_acm_certificate_validation" "example" {
  certificate_arn         = aws_acm_certificate.example.arn
  validation_record_fqdns = [for record in aws_route53_record.example : record.fqdn]
}

resource "aws_lb_listener" "example" {
  # ... other configuration ...

  certificate_arn = aws_acm_certificate_validation.example.certificate_arn
}
```

### Alternative Domains DNS Validation with Route 53

```hcl
resource "aws_acm_certificate" "example" {
  domain_name               = "example.com"
  subject_alternative_names = ["www.example.com", "example.org"]
  validation_method         = "DNS"
}

data "aws_route53_zone" "example_com" {
  name         = "example.com"
  private_zone = false
}

data "aws_route53_zone" "example_org" {
  name         = "example.org"
  private_zone = false
}

resource "aws_route53_record" "example" {
  for_each = {
    for dvo in aws_acm_certificate.example.domain_validation_options : dvo.domain_name => {
      name    = dvo.resource_record_name
      record  = dvo.resource_record_value
      type    = dvo.resource_record_type
      zone_id = dvo.domain_name == "example.org" ? data.aws_route53_zone.example_org.zone_id : data.aws_route53_zone.example_com.zone_id
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = each.value.zone_id
}

resource "aws_acm_certificate_validation" "example" {
  certificate_arn         = aws_acm_certificate.example.arn
  validation_record_fqdns = [for record in aws_route53_record.example : record.fqdn]
}

resource "aws_lb_listener" "example" {
  # ... other configuration ...

  certificate_arn = aws_acm_certificate_validation.example.certificate_arn
}
```

### Email Validation

In this situation, the resource is simply a waiter for manual email approval of ACM certificates.

```hcl
resource "aws_acm_certificate" "example" {
  domain_name       = "example.com"
  validation_method = "EMAIL"
}

resource "aws_acm_certificate_validation" "example" {
  certificate_arn = aws_acm_certificate.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `certificate_arn` - (Required) The ARN of the certificate that is being validated.
* `validation_record_fqdns` - (Optional) List of FQDNs that implement the validation. Only valid for DNS validation method ACM certificates. If this is set, the resource can implement additional sanity checks and has an explicit dependency on the resource that is implementing the validation

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The time at which the certificate was issued

## Timeouts

`acm_certificate_validation` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `create` - (Default `45m`) How long to wait for a certificate to be issued.
