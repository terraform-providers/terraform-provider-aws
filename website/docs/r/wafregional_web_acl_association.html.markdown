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

resource "aws_api_gateway_rest_api" "example" {
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/path1" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })

  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id

  triggers = {
    redeployment = sha1(jsonencode(aws_api_gateway_rest_api.example.body))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "example" {
  deployment_id = aws_api_gateway_deployment.example.id
  rest_api_id   = aws_api_gateway_rest_api.example.id
  stage_name    = "example"
}

resource "aws_wafregional_web_acl_association" "association" {
  resource_arn = aws_api_gateway_stage.example.arn
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
