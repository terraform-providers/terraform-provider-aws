---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_domain"
description: |-
  Provides an Lightsail Domain
---

# Resource: aws_lightsail_domain

Creates a domain resource for the specified domain (e.g., example.com).
You cannot register a new domain name using Lightsail. You must register
a domain name using Amazon Route 53 or another domain name registrar.
If you have already registered your domain, you can enter its name in
this parameter to manage the DNS records for that domain.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage, creating a new domain

```hcl
resource "aws_lightsail_domain" "domain_test" {
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
