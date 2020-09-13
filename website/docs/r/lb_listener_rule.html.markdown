---
subcategory: "Elastic Load Balancing v2 (ALB/NLB)"
layout: "aws"
page_title: "AWS: aws_lb_listener_rule"
description: |-
  Provides a Load Balancer Listener Rule resource.
---

# Resource: aws_lb_listener_rule

Provides a Load Balancer Listener Rule resource.

~> **Note:** `aws_alb_listener_rule` is known as `aws_lb_listener_rule`. The functionality is identical.

## Example Usage

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  # Other parameters
}

resource "aws_lb_listener_rule" "static" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }

  condition {
    host_header {
      values = ["example.com"]
    }
  }
}

# Forward action

resource "aws_lb_listener_rule" "host_based_weighted_routing" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 99

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }

  condition {
    host_header {
      values = ["my-service.*.terraform.io"]
    }
  }
}

# Weighted Forward action

resource "aws_lb_listener_rule" "host_based_routing" {
  listener_arn = aws_lb_listener.front_end.arn
  priority     = 99

  action {
    type = "forward"
    forward {
      target_group {
        arn    = aws_lb_target_group.main.arn
        weight = 80
      }

      target_group {
        arn    = aws_lb_target_group.canary.arn
        weight = 20
      }

      stickiness {
        enabled  = true
        duration = 600
      }
    }
  }

  condition {
    host_header {
      values = ["my-service.*.terraform.io"]
    }
  }
}

# Redirect action

resource "aws_lb_listener_rule" "redirect_http_to_https" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  condition {
    http_header {
      http_header_name = "X-Forwarded-For"
      values           = ["192.168.1.*"]
    }
  }
}

# Fixed-response action

resource "aws_lb_listener_rule" "health_check" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "HEALTHY"
      status_code  = "200"
    }
  }

  condition {
    query_string {
      key   = "health"
      value = "check"
    }

    query_string {
      value = "bar"
    }
  }
}

# Authenticate-cognito Action

resource "aws_cognito_user_pool" "pool" {
  # ...
}

resource "aws_cognito_user_pool_client" "client" {
  # ...
}

resource "aws_cognito_user_pool_domain" "domain" {
  # ...
}

resource "aws_lb_listener_rule" "admin" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
    type = "authenticate-cognito"

    authenticate_cognito {
      user_pool_arn       = aws_cognito_user_pool.pool.arn
      user_pool_client_id = aws_cognito_user_pool_client.client.id
      user_pool_domain    = aws_cognito_user_pool_domain.domain.domain
    }
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }
}

# Authenticate-oidc Action

resource "aws_lb_listener_rule" "oidc" {
  listener_arn = aws_lb_listener.front_end.arn

  action {
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

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.static.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `listener_arn` - (Required, Forces New Resource) The ARN of the listener to which to attach the rule.
* `priority` - (Optional) The priority for the rule between `1` and `50000`. Leaving it unset will automatically set the rule with next available priority after currently existing highest rule. A listener can't have multiple rules with the same priority.
* `action` - (Required) An Action block. Action blocks are documented below.
* `condition` - (Required) A Condition block. Multiple condition blocks of different types can be set and all must be satisfied for the rule to match. Condition blocks are documented below.

### Action Blocks

Action Blocks (for `action`) support the following:

* `type` - (Required) The type of routing action. Valid values are `forward`, `redirect`, `fixed-response`, `authenticate-cognito` and `authenticate-oidc`.
* `target_group_arn` - (Optional) The ARN of the Target Group to which to route traffic. Specify only if `type` is `forward` and you want to route to a single target group. To route to one or more target groups, use a `forward` block instead.
* `forward` - (Optional) Information for creating an action that distributes requests among one or more target groups. Specify only if `type` is `forward`. If you specify both `forward` block and `target_group_arn` attribute, you can specify only one target group using `forward` and it must be the same target group specified in `target_group_arn`.
* `redirect` - (Optional) Information for creating a redirect action. Required if `type` is `redirect`.
* `fixed_response` - (Optional) Information for creating an action that returns a custom HTTP response. Required if `type` is `fixed-response`.
* `authenticate_cognito` - (Optional) Information for creating an authenticate action using Cognito. Required if `type` is `authenticate-cognito`.
* `authenticate_oidc` - (Optional) Information for creating an authenticate action using OIDC. Required if `type` is `authenticate-oidc`.

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

### Condition Blocks

One or more condition blocks can be set per rule. Most condition types can only be specified once per rule except for `http-header` and `query-string` which can be specified multiple times.

Condition Blocks (for `condition`) support the following:

* `host_header` - (Optional) Contains a single `values` item which is a list of host header patterns to match. The maximum size of each pattern is 128 characters. Comparison is case insensitive. Wildcard characters supported: * (matches 0 or more characters) and ? (matches exactly 1 character). Only one pattern needs to match for the condition to be satisfied.
* `http_header` - (Optional) HTTP headers to match. [HTTP Header block](#http-header-blocks) fields documented below.
* `http_request_method` - (Optional) Contains a single `values` item which is a list of HTTP request methods or verbs to match. Maximum size is 40 characters. Only allowed characters are A-Z, hyphen (-) and underscore (\_). Comparison is case sensitive. Wildcards are not supported. Only one needs to match for the condition to be satisfied. AWS recommends that GET and HEAD requests are routed in the same way because the response to a HEAD request may be cached.
* `path_pattern` - (Optional) Contains a single `values` item which is a list of path patterns to match against the request URL. Maximum size of each pattern is 128 characters. Comparison is case sensitive. Wildcard characters supported: * (matches 0 or more characters) and ? (matches exactly 1 character). Only one pattern needs to match for the condition to be satisfied. Path pattern is compared only to the path of the URL, not to its query string. To compare against the query string, use a `query_string` condition.
* `query_string` - (Optional) Query strings to match. [Query String block](#query-string-blocks) fields documented below.
* `source_ip` - (Optional) Contains a single `values` item which is a list of source IP CIDR notations to match. You can use both IPv4 and IPv6 addresses. Wildcards are not supported. Condition is satisfied if the source IP address of the request matches one of the CIDR blocks. Condition is not satisfied by the addresses in the `X-Forwarded-For` header, use `http_header` condition instead.

~> **NOTE::** Exactly one of `host_header`, `http_header`, `http_request_method`, `path_pattern`, `query_string` or `source_ip` must be set per condition.

#### HTTP Header Blocks

HTTP Header Blocks (for `http_header`) support the following:

* `http_header_name` - (Required) Name of HTTP header to search. The maximum size is 40 characters. Comparison is case insensitive. Only RFC7240 characters are supported. Wildcards are not supported. You cannot use HTTP header condition to specify the host header, use a `host-header` condition instead.
* `values` - (Required) List of header value patterns to match. Maximum size of each pattern is 128 characters. Comparison is case insensitive. Wildcard characters supported: * (matches 0 or more characters) and ? (matches exactly 1 character). If the same header appears multiple times in the request they will be searched in order until a match is found. Only one pattern needs to match for the condition to be satisfied. To require that all of the strings are a match, create one condition block per string.

#### Query String Blocks

Query String Blocks (for `query_string`) support the following:

* `values` - (Required) Query string pairs or values to match. Query String Value blocks documented below. Multiple `values` blocks can be specified, see example above. Maximum size of each string is 128 characters. Comparison is case insensitive. Wildcard characters supported: * (matches 0 or more characters) and ? (matches exactly 1 character). To search for a literal '\*' or '?' character in a query string, escape the character with a backslash (\\). Only one pair needs to match for the condition to be satisfied.

Query String Value Blocks (for `query_string.values`) support the following:

* `key` - (Optional) Query string key pattern to match.
* `value` - (Required) Query string value pattern to match.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the rule (matches `arn`)
* `arn` - The ARN of the rule (matches `id`)

## Import

Rules can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener_rule.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener-rule/app/test/8e4497da625e2d8a/9ab28ade35828f96/67b3d2d36dd7c26b
```
