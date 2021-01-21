---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb_listener"
description: |-
  Provides a Load Balancer Listener resource.
---

# Resource: aws_lb_listener

Provides a Load Balancer Listener resource.

~> **Note:** `aws_alb_listener` is known as `aws_lb_listener`. The functionality is identical.

## Example Usage

### Forward Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_target_group" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "arn:aws:iam::187416307283:server-certificate/test_cert_rab3wuqwgja25ct3n4jdj2tzu4"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.front_end.arn
  }
}
```

### Redirect Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

### Fixed-response Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code  = "200"
    }
  }
}
```

### Authenticate-cognito Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_target_group" "front_end" {
  # ...
}

resource "aws_cognito_user_pool" "pool" {
  # ...
}

resource "aws_cognito_user_pool_client" "client" {
  # ...
}

resource "aws_cognito_user_pool_domain" "domain" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "authenticate-cognito"

    authenticate_cognito {
      user_pool_arn       = aws_cognito_user_pool.pool.arn
      user_pool_client_id = aws_cognito_user_pool_client.client.id
      user_pool_domain    = aws_cognito_user_pool_domain.domain.domain
    }
  }

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.front_end.arn
  }
}
```

### Authenticate-oidc Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_target_group" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.front_end.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "client_id"
      client_secret          = "client_secret"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"
    }
  }

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.front_end.arn
  }
}
```

### Gateway Load Balancer Listener

```hcl
resource "aws_lb" "example" {
  load_balancer_type = "gateway"
  name               = "example"

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }
}

resource "aws_lb_target_group" "example" {
  name     = "example"
  port     = 6081
  protocol = "GENEVE"
  vpc_id   = aws_vpc.example.id

  health_check {
    port     = 80
    protocol = "HTTP"
  }
}

resource "aws_lb_listener" "example" {
  load_balancer_arn = aws_lb.example.id

  default_action {
    target_group_arn = aws_lb_target_group.example.id
    type             = "forward"
  }
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_arn` - (Required, Forces New Resource) The ARN of the load balancer.
* `port` - (Optional) The port on which the load balancer is listening. Not valid for Gateway Load Balancers.
* `protocol` - (Optional) The protocol for connections from clients to the load balancer. For Application Load Balancers, valid values are `HTTP` and `HTTPS`, with a default of `HTTP`. For Network Load Balancers, valid values are `TCP`, `TLS`, `UDP`, and `TCP_UDP`. Not valid to use `UDP` or `TCP_UDP` if dual-stack mode is enabled. Not valid for Gateway Load Balancers.
* `ssl_policy` - (Optional) The name of the SSL Policy for the listener. Required if `protocol` is `HTTPS` or `TLS`.
* `certificate_arn` - (Optional) The ARN of the default SSL server certificate. Exactly one certificate is required if the protocol is HTTPS. For adding additional SSL certificates, see the [`aws_lb_listener_certificate` resource](/docs/providers/aws/r/lb_listener_certificate.html).
* `default_action` - (Required) An Action block. Action blocks are documented below.

~> **NOTE::** Please note that listeners that are attached to Application Load Balancers must use either `HTTP` or `HTTPS` protocols while listeners that are attached to Network Load Balancers must use the `TCP` protocol.

Action Blocks (for `default_action`) support the following:

* `type` - (Required) The type of routing action. Valid values are `forward`, `redirect`, `fixed-response`, `authenticate-cognito` and `authenticate-oidc`.
* `target_group_arn` - (Optional) The ARN of the Target Group to which to route traffic. Specify only if `type` is `forward` and you want to route to a single target group. To route to one or more target groups, use a `forward` block instead.
* `forward` - (Optional) Information for creating an action that distributes requests among one or more target groups. Specify only if `type` is `forward`. If you specify both `forward` block and `target_group_arn` attribute, you can specify only one target group using `forward` and it must be the same target group specified in `target_group_arn`.
* `redirect` - (Optional) Information for creating a redirect action. Required if `type` is `redirect`.
* `fixed_response` - (Optional) Information for creating an action that returns a custom HTTP response. Required if `type` is `fixed-response`.

Forward Blocks (for `forward`) support the following:

* `target_group` - (Required) One or more target groups block.
* `stickiness` - (Optional) The target group stickiness for the rule.

Target Group Blocks (for `target_group`) supports the following:

* `arn` - (Required) The Amazon Resource Name (ARN) of the target group.
* `weight` - (Optional) The weight. The range is 0 to 999.

Target Group Stickiness Config Blocks (for `stickiness`) supports the following:

* `enabled` - (Required) Indicates whether target group stickiness is enabled.
* `duration` - (Optional) The time period, in seconds, during which requests from a client should be routed to the same target group. The range is 1-604800 seconds (7 days).


Redirect Blocks (for `redirect`) support the following:

~> **NOTE::** You can reuse URI components using the following reserved keywords: `#{protocol}`, `#{host}`, `#{port}`, `#{path}` (the leading "/" is removed) and `#{query}`.

* `host` - (Optional) The hostname. This component is not percent-encoded. The hostname can contain `#{host}`. Defaults to `#{host}`.
* `path` - (Optional) The absolute path, starting with the leading "/". This component is not percent-encoded. The path can contain #{host}, #{path}, and #{port}. Defaults to `/#{path}`.
* `port` - (Optional) The port. Specify a value from `1` to `65535` or `#{port}`. Defaults to `#{port}`.
* `protocol` - (Optional) The protocol. Valid values are `HTTP`, `HTTPS`, or `#{protocol}`. Defaults to `#{protocol}`.
* `query` - (Optional) The query parameters, URL-encoded when necessary, but not percent-encoded. Do not include the leading "?". Defaults to `#{query}`.
* `status_code` - (Required) The HTTP redirect code. The redirect is either permanent (`HTTP_301`) or temporary (`HTTP_302`).

Fixed-response Blocks (for `fixed_response`) support the following:

* `content_type` - (Required) The content type. Valid values are `text/plain`, `text/css`, `text/html`, `application/javascript` and `application/json`.
* `message_body` - (Optional) The message body.
* `status_code` - (Optional) The HTTP response code. Valid values are `2XX`, `4XX`, or `5XX`.

Authenticate Cognito Blocks (for `authenticate_cognito`) supports the following:

* `authentication_request_extra_params` - (Optional) The query parameters to include in the redirect request to the authorization endpoint. Max: 10.
* `on_unauthenticated_request` - (Optional) The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`
* `scope` - (Optional) The set of user claims to be requested from the IdP.
* `session_cookie_name` - (Optional) The name of the cookie used to maintain session information.
* `session_timeout` - (Optional) The maximum duration of the authentication session, in seconds.
* `user_pool_arn` - (Required) The ARN of the Cognito user pool.
* `user_pool_client_id` - (Required) The ID of the Cognito user pool client.
* `user_pool_domain` - (Required) The domain prefix or fully-qualified domain name of the Cognito user pool.

Authenticate OIDC Blocks (for `authenticate_oidc`) supports the following:

* `authentication_request_extra_params` - (Optional) The query parameters to include in the redirect request to the authorization endpoint. Max: 10.
* `authorization_endpoint` - (Required) The authorization endpoint of the IdP.
* `client_id` - (Required) The OAuth 2.0 client identifier.
* `client_secret` - (Required) The OAuth 2.0 client secret.
* `issuer` - (Required) The OIDC issuer identifier of the IdP.
* `on_unauthenticated_request` - (Optional) The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`
* `scope` - (Optional) The set of user claims to be requested from the IdP.
* `session_cookie_name` - (Optional) The name of the cookie used to maintain session information.
* `session_timeout` - (Optional) The maximum duration of the authentication session, in seconds.
* `token_endpoint` - (Required) The token endpoint of the IdP.
* `user_info_endpoint` - (Required) The user info endpoint of the IdP.

Authentication Request Extra Params Blocks (for `authentication_request_extra_params`) supports the following:

* `key` - (Required) The key of query parameter
* `value` - (Required) The value of query parameter

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the listener (matches `arn`)
* `arn` - The ARN of the listener (matches `id`)

## Import

Listeners can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener/app/front-end-alb/8e4497da625e2d8a/9ab28ade35828f96
```
