---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_endpoint"
description: |-
    Provides details about a specific Route 53 Resolver Endpoint
---

# Data Source: aws_route53_resolver_endpoint

`aws_route53_resolver_endpoint` provides details about a specific Route53 Resolver Endpoint.

This data source allows to find a list of IPaddresses associated with a specific Route53 Resolver Endpoint.

## Example Usage

```hcl
data "aws_route53_resolver_endpoint" "example" {
  resolver_endpoint_id = "rslvr-in-1abc2345ef678g91h"
}
```

```hcl
data "aws_route53_resolver_endpoint" "example" {
  filter {
    name   = "NAME"
    values = ["MyResolverExampleName"]
  }
}
```

## Argument Reference

* `resolver_endpoint_id` - (Optional) The ID of the Route53 Resolver Endpoint.
* `filter` - (Optional) One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[Route53resolver Filter value in the AWS API reference][1].

In addition to all arguments above, the following attributes are exported:

* `arn` - The computed ARN of the Route53 Resolver Endpoint.
* `direction` - The direction of the queries to or from the Resolver Endpoint .
* `ip_addresses` - A list of IPaddresses that have been associated with the Resolver Endpoint.
* `status` - The current status of the Resolver Endpoint.
* `vpc_id` - The ID of the Host VPC that the Resolver Endpoint resides in.

[1]: https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53resolver_Filter.html