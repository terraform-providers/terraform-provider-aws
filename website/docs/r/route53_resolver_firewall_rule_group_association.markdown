---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group_association"
description: |-
  Provides a Route 53 Resolver DNS Firewall rule group association resource.
---

# Resource: aws_route53_resolver_firewall_rule_group_association

Provides a Route 53 Resolver DNS Firewall rule group association resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_rule_group" "example" {
  name = "example"
}

resource "aws_route53_resolver_firewall_rule_group_association" "example" {
  name                   = "example"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.example.id
  priority               = 100
  vpc_id                 = aws_vpc.example.id
}
```

## Argument Reference

The following argument is supported:

* `name` - (Required) A name that lets you identify the rule group association, to manage and use it.
* `firewall_rule_group_id` - (Required) The unique identifier of the firewall rule group.
* `mutation_protection` - (Optional) If enabled, this setting disallows modification or removal of the association, to help prevent against accidentally altering DNS firewall protections. Valid values: `ENABLED`, `DISABLED`.
* `priority` - (Required) The setting that determines the processing order of the rule group among the rule groups that you associate with the specified VPC. DNS Firewall filters VPC traffic starting from the rule group with the lowest numeric priority setting.
* `vpc_id` - (Required) The unique identifier of the VPC that you want to associate with the rule group.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN (Amazon Resource Name) of the firewall rule group association.
* `id` - The identifier for the association.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Route 53 Resolver DNS Firewall rule group associations can be imported using the Route 53 Resolver DNS Firewall rule group association ID, e.g.

```
$ terraform import aws_route53_resolver_firewall_rule_group_association.example rslvr-frgassoc-0123456789abcdef
```
