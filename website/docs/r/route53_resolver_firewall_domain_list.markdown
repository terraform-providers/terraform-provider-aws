---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_domain_list"
description: |-
  Provides a Route 53 Resolver DNS Firewall domain list resource.
---

# Resource: aws_route53_resolver_firewall_domain_list

Provides a Route 53 Resolver DNS Firewall domain list resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_domain_list" "example" {
  name = "example"
}
```

## Argument Reference

The following argument is supported:

* `name` - (Required) A name that lets you identify the domain list, to manage and use it.
* `domains` - (Optional) A array of domains for the firewall domain list.
* `tags` - (Optional) A map of tags to assign to the resource. f configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN (Amazon Resource Name) of the domain list.
* `id` - The ID of the domain list.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

 Route 53 Resolver DNS Firewall domain lists can be imported using the Route 53 Resolver DNS Firewall domain list ID, e.g.

```
$ terraform import aws_route53_resolver_firewall_domain_list.example rslvr-fdl-0123456789abcdef
```
