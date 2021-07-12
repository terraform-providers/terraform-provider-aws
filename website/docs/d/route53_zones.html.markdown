---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_route53_zones"
description: |-
    Provides a list of Route53 Hosted Zone ids in a region
---

# Data Source: aws_route53_zones

This resource can be useful for getting back a list of Route53 Hosted Zone ids for a region.


## Example Usage

The following example retrieves a list of all Hosted Zone ids.

```hcl
data "aws_route53_zones" "all" {}

output "foo" {
  value = data.aws_route53_zones.all.ids
}
```

## Attributes Reference

* `ids` - A list of all the Route53 Hosted Zone Ids found. This data source will fail if none are found.
