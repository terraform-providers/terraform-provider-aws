---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl_association"
description: |-
  Manages an association with WAF Regional Web ACL
---

# Resource: aws_wafregional_web_acl_association

Manages an association with WAF Regional Web ACL.

-> **Note:** An Application Load Balancer can only be associated with one WAF Regional WebACL.

## Application Load Balancer Association Example

```hcl
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "foo" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_wafregional_web_acl" "foo" {
  name        = "foo"
  metric_name = "foo"

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.foo.id
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "bar" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

resource "aws_alb" "foo" {
  internal = true
  subnets  = [aws_subnet.foo.id, aws_subnet.bar.id]
}

resource "aws_wafregional_web_acl_association" "foo" {
  resource_arn = aws_alb.foo.arn
  web_acl_id   = aws_wafregional_web_acl.foo.id
}
```

## API Gateway Association Example

```hcl
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "foo" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_wafregional_web_acl" "foo" {
  name        = "foo"
  metric_name = "foo"

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.foo.id
  }
}

resource "aws_api_gateway_rest_api" "test" {
  name = "foo"
}

resource "aws_api_gateway_resource" "test" {
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method" "test" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method_response" "test" {
  http_method = aws_api_gateway_method.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "GET"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  type                    = "HTTP"
  uri                     = "http://www.example.com"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration_response.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "test"
}

resource "aws_wafregional_web_acl_association" "association" {
  resource_arn = aws_api_gateway_stage.test.arn
  web_acl_id   = aws_wafregional_web_acl.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `web_acl_id` - (Required) The ID of the WAF Regional WebACL to create an association.
* `resource_arn` - (Required) ARN of the resource to associate with. For example, an Application Load Balancer or API Gateway Stage.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association

## Import

WAF Regional Web ACL Association can be imported using their `web_acl_id:resource_arn`, e.g.

```
$ terraform import aws_wafregional_web_acl_association.foo web_acl_id:resource_arn
```
