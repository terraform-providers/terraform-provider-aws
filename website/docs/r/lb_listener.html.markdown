---
layout: "aws"
page_title: "AWS: aws_lb_listener"
sidebar_current: "docs-aws-resource-elbv2-listener"
description: |-
  Provides a Load Balancer Listener resource.
---

# aws_lb_listener

Provides a Load Balancer Listener resource.

~> **Note:** `aws_alb_listener` is known as `aws_lb_listener`. The functionality is identical.

## Example Usage

```hcl
# Create a new load balancer
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_target_group" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.front_end.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "arn:aws:iam::187416307283:server-certificate/test_cert_rab3wuqwgja25ct3n4jdj2tzu4"

  default_action {
    target_group_arn = "${aws_lb_target_group.front_end.arn}"
    type             = "forward"
  }
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_arn` - (Required, Forces New Resource) The ARN of the load balancer.
* `port` - (Required) The port on which the load balancer is listening.
* `protocol` - (Optional) The protocol for connections from clients to the load balancer. Valid values are `TCP`, `HTTP` and `HTTPS`. Defaults to `HTTP`.
* `ssl_policy` - (Optional) The name of the SSL Policy for the listener. Required if `protocol` is `HTTPS`.
* `certificate_arn` - (Optional) The ARN of the default SSL server certificate. Exactly one certificate is required if the protocol is HTTPS. For adding additional SSL certificates, see the [`aws_lb_listener_certificate` resource](/docs/providers/aws/r/lb_listener_certificate.html).
* `default_action` - (Required) An Action block. Action blocks are documented below.
* `wait_for_capacity_timeout` (Default: "10m") A maximum
  [duration](https://golang.org/pkg/time/#ParseDuration) that Terraform should
  wait for Target Group instances to be healthy before timing out.  (See also [Waiting
  for Capacity](#waiting-for-capacity) below.) Setting this to "0" causes
  Terraform to skip Capacity Waiting behavior for associated Target Groups.
* `min_target_group_capacity` - (Optional) Setting this causes Terraform to wait for
  this number of instances to show up healthy in the Target Group associated with
  the `default_action` before adding or updating the default action to point to this
  target group. If not specified, this value will be derived from the "Desired" capacity
  configured on any ASGs targeted by the associated Target Group. 
  (See also [Waiting for Capacity](#waiting-for-capacity) below.)

~> **NOTE::** Please note that listeners that are attached to Application Load Balancers must use either `HTTP` or `HTTPS` protocols while listeners that are attached to Network Load Balancers must use the `TCP` protocol.

Action Blocks (for `default_action`) support the following:

* `target_group_arn` - (Required) The ARN of the Target Group to which to route traffic.
* `type` - (Required) The type of routing action. The only valid value is `forward`.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the listener (matches `arn`)
* `arn` - The ARN of the listener (matches `id`)

## Waiting for Capacity

A newly-created Target Group is initially empty, with it's eventual members usually
composed of ASG instances that initialize as the ASG comes up to it's desired
capacity. These instances take time to launch and boot.

When creating or modifying Listeners on Load Balancers which route to these targets
(specifically when changing the default action), changes made prematurely can result
in connections routing to instances that are not (yet) ready to serve requests.

This is complicated by the fact that Target Group health status will is reported as
`unused` until the Target Group is actually attached to a Listener.

Terraform provides a mechanism to help consistently manage Listener modifications
with respect to Target Groups and ASG scale up time across dependent resources.

#### Waiting for Target Group Capacity

Terraform (by default) waits upon Listener creation for `Desired` healthy instances
to show up in the associated Target Group before continuing
(_where `Desired` is read from the ASG targeted by the associated Target Group_).

Terraform will also wait (by default) upon Listener modification of the default action
for `Desired` healthy instances to show up in the associated Target Group before
changing the Listener's default action to route to that Target Group.

You can choose a different minimum number of instances (than `Desired` above) by specifying
the `min_target_group_capacity` attribute.

Terraform considers an instance "healthy" when the Target Group reports status `healthy`
for that instance. See the [AWS Target Group Health Check Docs](http://docs.aws.amazon.com/elasticloadbalancing/latest/application/target-group-health-checks.html)
for more information on Target Group health checks.

Terraform will wait for healthy instances for up to
`wait_for_capacity_timeout`. If Target Group health is taking more than a few minutes
to reach the minimum desired capacity, it's worth investigating for scaling activity
errors, which can be caused by problems with the Launch Configuration of the associated ASGs.

Setting `wait_for_capacity_timeout` to `"0"` disables Target Group health Capacity waiting.

_Terraform accomplishes this wait by creating a temporary rule on the associated listener which
routes to the awaited Target Group; the rule is a `host-header` rule which uses a value of
`health-check.terraform.localhost` (chosen because it is highly unlikely to interfere with any
host headers used in normal application operation). This allows the Target Group's health to be
checked by the Load Balancer while protecting it from receiving traffic until the desired health
status is achieved._

#### Troubleshooting Capacity Waiting Timeouts

If ASG creation takes more than a few minutes, this could indicate one of a
number of configuration problems. See the [AWS Docs on Load Balancer
Troubleshooting](https://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/elb-troubleshooting.html)
for more information.

## Import

Listeners can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener/app/front-end-alb/8e4497da625e2d8a/9ab28ade35828f96
```
