---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_query_log_config_association"
description: |-
  Provides a Route 53 Resolver query logging configuration association resource.
---

# Resource: aws_route53_resolver_query_log_config_association

Provides a Route 53 Resolver query logging configuration association resource.

## Example Usage

```hcl
resource "aws_route53_resolver_query_log_config_association" "example" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.example.id
  resource_id                  = aws_vpc.example.id
}
```

## Argument Reference

The following arguments are supported:

* `resolver_query_log_config_id` - (Required) The ID of the [Route 53 Resolver query logging configuration](route53_resolver_query_log_config.html) that you want to associate a VPC with.
* `resource_id` - (Required) The ID of a VPC that you want this query logging configuration to log queries for.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` -The ID of the Route 53 Resolver query logging configuration association.

## Import

 Route 53 Resolver query logging configuration associations can be imported using the Route 53 Resolver query logging configuration association ID, e.g.

```
$ terraform import aws_route53_resolver_query_log_config_association.example rqlca-b320624fef3c4d70
```
