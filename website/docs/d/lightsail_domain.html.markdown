---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain"
description: |-
  Provides details about a Lightsail Domain
---

# Data Source: aws_lightsail_domain

Provides details about a Lightsail Domain for the specified domain (e.g., example.com).

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage, getting details of a domain

```hcl
data "aws_lightsail_domain" "domain_test" {
  domain_name = "mydomain.com"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of the Lightsail domain to manage

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name used for this domain
* `arn` - The ARN of the Lightsail domain
* `tags` - A map of tags assigned to the resource.