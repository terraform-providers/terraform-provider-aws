---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_dnssec_config"
description: |-
  Provides a Route 53 Resolver DNSSEC config resource.
---

# Resource: aws_route53_resolver_dnssec_config

Provides a Route 53 Resolver DNSSEC config resource.

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_route53_resolver_dnssec_config" "example" {
  resource_id = aws_vpc.example.id
}
```

## Argument Reference

The following argument is supported:

* `resource_id` - (Required) The ID of the virtual private cloud (VPC) that you're updating the DNSSEC validation status for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN for a configuration for DNSSEC validation.
* `id` - The ID for a configuration for DNSSEC validation.
* `owner_id` - The owner account ID of the virtual private cloud (VPC) for a configuration for DNSSEC validation.
* `validation_status` - The validation status for a DNSSEC configuration. The status can be one of the following: `ENABLING`, `ENABLED`, `DISABLING` and `DISABLED`.

## Import

 Route 53 Resolver DNSSEC configs can be imported using the Route 53 Resolver DNSSEC config ID, e.g.

```
$ terraform import aws_route53_resolver_dnssec_config.example rdsc-be1866ecc1683e95
```
