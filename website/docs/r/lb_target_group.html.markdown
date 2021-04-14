---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb_target_group"
description: |-
  Provides a Target Group resource for use with Load Balancers.
---

# Resource: aws_lb_target_group

Provides a Target Group resource for use with Load Balancer resources.

~> **Note:** `aws_alb_target_group` is known as `aws_lb_target_group`. The functionality is identical.

## Example Usage

### Instance Target Group

```terraform
resource "aws_lb_target_group" "test" {
  name     = "tf-example-lb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### IP Target Group

```terraform
resource "aws_lb_target_group" "ip-example" {
  name        = "tf-example-lb-tg"
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.main.id
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}
```

### Lambda Target Group

```terraform
resource "aws_lb_target_group" "lambda-example" {
  name        = "tf-example-lb-tg"
  target_type = "lambda"
}
```

## Argument Reference

The following arguments are supported:

* `deregistration_delay` - (Optional) Amount time for Elastic Load Balancing to wait before changing the state of a deregistering target from draining to unused. The range is 0-3600 seconds. The default value is 300 seconds.
* `health_check` - (Optional, Maximum of 1) Health Check configuration block. Detailed below.
* `lambda_multi_value_headers_enabled` - (Optional) Whether the request and response headers exchanged between the load balancer and the Lambda function include arrays of values or strings. Only applies when `target_type` is `lambda`.
* `load_balancing_algorithm_type` - (Optional) Determines how the load balancer selects targets when routing requests. Only applicable for Application Load Balancer Target Groups. The value is `round_robin` or `least_outstanding_requests`. The default is `round_robin`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`. Cannot be longer than 6 characters.
* `name` - (Optional, Forces new resource) Name of the target group. If omitted, Terraform will assign a random, unique name.
* `port` - (May be required, Forces new resource) Port on which targets receive traffic, unless overridden when registering a specific target. Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.
* `preserve_client_ip` - (Optional) Whether client IP preservation is enabled. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#client-ip-preservation) for more information.
* `protocol_version` - (Optional, Forces new resource) Only applicable when `protocol` is `HTTP` or `HTTPS`. The protocol version. Specify GRPC to send requests to targets using gRPC. Specify HTTP2 to send requests to targets using HTTP/2. The default is HTTP1, which sends requests to targets using HTTP/1.1
* `protocol` - (May be required, Forces new resource) Protocol to use for routing traffic to the targets. Should be one of `GENEVE`, `HTTP`, `HTTPS`, `TCP`, `TCP_UDP`, `TLS`, or `UDP`. Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.
* `proxy_protocol_v2` - (Optional) Whether to enable support for proxy protocol v2 on Network Load Balancers. See [doc](https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-target-groups.html#proxy-protocol) for more information. Default is `false`.
* `slow_start` - (Optional) Amount time for targets to warm up before the load balancer sends them a full share of requests. The range is 30-900 seconds or 0 to disable. The default value is 0 seconds.
* `stickiness` - (Optional, Maximum of 1) Stickiness configuration block. Detailed below.
* `tags` - (Optional) Map of tags to assign to the resource.
* `target_type` - (May be required, Forces new resource) Type of target that you must specify when registering targets with this target group. The possible values are `instance` (targets are specified by instance ID) or `ip` (targets are specified by IP address) or `lambda` (targets are specified by lambda arn). The default is `instance`. Note that you can't specify targets for a target group using both instance IDs and IP addresses. If the target type is `ip`, specify IP addresses from the subnets of the virtual private cloud (VPC) for the target group, the RFC 1918 range (10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16), and the RFC 6598 range (100.64.0.0/10). You can't specify publicly routable IP addresses.
* `vpc_id` - (Optional, Forces new resource) Identifier of the VPC in which to create the target group. Required when `target_type` is `instance` or `ip`. Does not apply when `target_type` is `lambda`.

### health_check

~> **Note:** The Health Check parameters you can set vary by the `protocol` of the Target Group. Many parameters cannot be set to custom values for `network` load balancers at this time. See http://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html for a complete reference. Keep in mind, that health checks produce actual requests to the backend. The underlying function is invoked when `target_type` is set to `lambda`.

* `enabled` - (Optional) Whether health checks are enabled. Defaults to `true`.
* `healthy_threshold` - (Optional) Number of consecutive health checks successes required before considering an unhealthy target healthy. Defaults to 3.
* `interval` - (Optional) Approximate amount of time, in seconds, between health checks of an individual target. Minimum value 5 seconds, Maximum value 300 seconds. For `lambda` target groups, it needs to be greater as the `timeout` of the underlying `lambda`. Default 30 seconds.
* `matcher` (May be required) Response codes to use when checking for a healthy responses from a target. You can specify multiple values (for example, "200,202" for HTTP(s) or "0,12" for GRPC) or a range of values (for example, "200-299" or "0-99"). Required for HTTP/HTTPS/GRPC ALB. Only applies to Application Load Balancers (i.e., HTTP/HTTPS/GRPC) not Network Load Balancers (i.e., TCP).
* `path` - (May be required) Destination for the health check request. Required for HTTP/HTTPS ALB and HTTP NLB. Only applies to HTTP/HTTPS.
* `port` - (Optional) Port to use to connect with the target. Valid values are either ports 1-65535, or `traffic-port`. Defaults to `traffic-port`.
* `protocol_version` - (Optional) Protocol version. Defaults to `HTTP1`. Specify GRPC to send requests to targets using GRPC, HTTP2 to send requests to targets using HTTP/2, HTTP1 to send requests to targets using HTTP/1.1.
* `protocol` - (Optional) Protocol to use to connect with the target. Defaults to `HTTP`. Not applicable when `target_type` is `lambda`.
* `timeout` - (Optional) Amount of time, in seconds, during which no response means a failed health check. For Application Load Balancers, the range is 2 to 120 seconds, and the default is 5 seconds for the `instance` target type and 30 seconds for the `lambda` target type. For Network Load Balancers, you cannot set a custom value, and the default is 10 seconds for TCP and HTTPS health checks and 6 seconds for HTTP health checks.
* `unhealthy_threshold` - (Optional) Number of consecutive health check failures required before considering the target unhealthy. For Network Load Balancers, this value must be the same as the `healthy_threshold`. Defaults to 3.

### stickiness

~> **NOTE:** Currently, an NLB (i.e., protocol of `HTTP` or `HTTPS`) can have an invalid `stickiness` block with `type` set to `lb_cookie` as long as `enabled` is set to `false`. However, please update your configurations to avoid errors in a future version of the provider: either remove the invalid `stickiness` block or set the `type` to `source_ip`.

* `cookie_duration` - (Optional) Only used when the type is `lb_cookie`. The time period, in seconds, during which requests from a client should be routed to the same target. After this time period expires, the load balancer-generated cookie is considered stale. The range is 1 second to 1 week (604800 seconds). The default value is 1 day (86400 seconds).
* `enabled` - (Optional) Whether to enable `stickiness`. Default is `true`.
* `type` - (Required) Type of sticky sessions. The only current possible values are `lb_cookie` for ALBs and `source_ip` for NLBs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn_suffix` - ARN suffix for use with CloudWatch Metrics.
* `arn` - ARN of the Target Group (matches `id`).
* `id` - ARN of the Target Group (matches `arn`).
* `name` - Name of the Target Group.

## Import

Target Groups can be imported using their ARN, e.g.

```
$ terraform import aws_lb_target_group.app_front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:targetgroup/app-front-end/20cfe21448b66314
```
