---
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl_association"
sidebar_current: "docs-aws-resource-wafregional-web-acl-association"
description: |-
  Provides a resource to create an association between a WAF Regional WebACL and Application Load Balancer.
---

# aws_wafregional_web_acl_association

Provides a resource to create an association between a WAF Regional WebACL and Application Load Balancer.

-> **Note:** An Application Load Balancer can only be associated with one WAF Regional WebACL.

## Example Usage

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
    data_id = "${aws_wafregional_ipset.ipset.id}"
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
    rule_id  = "${aws_wafregional_rule.foo.id}"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "foo" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "10.1.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "bar" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "10.1.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_alb" "foo" {
  internal = true
  subnets  = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_wafregional_web_acl_association" "foo" {
  resource_arn = "${aws_alb.foo.arn}"
  web_acl_id   = "${aws_wafregional_web_acl.foo.id}"
}
```

## Argument Reference

The following arguments are supported:

* `web_acl_id` - (Required) The ID of the WAF Regional WebACL to create an association.
* `resource_arn` - (Required) Application Load Balancer ARN to associate with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association
