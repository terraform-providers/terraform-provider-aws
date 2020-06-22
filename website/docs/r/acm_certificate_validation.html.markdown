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
resource "aws_acm_certificate" "cert" {
  domain_name       = "example.com"
  validation_method = "DNS"
}

data "aws_route53_zone" "zone" {
  name         = "example.com."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name    = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type    = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.zone_id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn         = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["${aws_route53_record.cert_validation.fqdn}"]
}

resource "aws_lb_listener" "front_end" {
  # [...]
  certificate_arn = "${aws_acm_certificate_validation.cert.certificate_arn}"
}
```

### Alternative Domains DNS Validation with Route 53

```hcl
resource "aws_acm_certificate" "cert" {
  domain_name               = "example.com"
  subject_alternative_names = ["www.example.com", "example.org"]
  validation_method         = "DNS"
}

data "aws_route53_zone" "zone" {
  name         = "example.com."
  private_zone = false
}

data "aws_route53_zone" "zone_alt" {
  name         = "example.org."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name    = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type    = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.zone_id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl     = 60
}

resource "aws_route53_record" "cert_validation_alt1" {
  name    = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_name}"
  type    = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.zone_id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.1.resource_record_value}"]
  ttl     = 60
}

resource "aws_route53_record" "cert_validation_alt2" {
  name    = "${aws_acm_certificate.cert.domain_validation_options.2.resource_record_name}"
  type    = "${aws_acm_certificate.cert.domain_validation_options.2.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone_alt.zone_id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.2.resource_record_value}"]
  ttl     = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"

  validation_record_fqdns = [
    "${aws_route53_record.cert_validation.fqdn}",
    "${aws_route53_record.cert_validation_alt1.fqdn}",
    "${aws_route53_record.cert_validation_alt2.fqdn}",
  ]
}

resource "aws_lb_listener" "front_end" {
  # [...]
  certificate_arn = "${aws_acm_certificate_validation.cert.certificate_arn}"
}
```

### Email Validation

In this situation, the resource is simply a waiter for manual email approval of ACM certificates.

```hcl
resource "aws_acm_certificate" "cert" {
  domain_name       = "example.com"
  validation_method = "EMAIL"
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
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

`acm_certificate_validation` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

- `create` - (Default `45m`) How long to wait for a certificate to be issued.
