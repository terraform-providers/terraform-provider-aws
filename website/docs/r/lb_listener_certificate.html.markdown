---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb_listener_certificate"
description: |-
  Provides a Load Balancer Listener Certificate resource.
---

# Resource: aws_lb_listener_certificate

Provides a Load Balancer Listener Certificate resource.

This resource is for additional certificates and does not replace the default certificate on the listener.

~> **Note:** `aws_alb_listener_certificate` is known as `aws_lb_listener_certificate`. The functionality is identical.

## Example Usage

```hcl
resource "aws_acm_certificate" "example" {
  # ...
}

resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  # ...
}

resource "aws_lb_listener_certificate" "example" {
  listener_arn    = aws_lb_listener.front_end.arn
  certificate_arn = aws_acm_certificate.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `listener_arn` - (Required, Forces New Resource) The ARN of the listener to which to attach the certificate.
* `certificate_arn` - (Required, Forces New Resource) The ARN of the certificate to attach to the listener.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `listener_arn` and `certificate_arn` separated by a `_`.

## Import

Listener Certificates can be imported using their id, e.g.

```
$ terraform import aws_lb_listener_certificate.example arn:aws:elasticloadbalancing:us-west-2:123456789012:listener/app/test/8e4497da625e2d8a/9ab28ade35828f96/67b3d2d36dd7c26b_arn:aws:iam::123456789012:server-certificate/tf-acc-test-6453083910015726063
```
