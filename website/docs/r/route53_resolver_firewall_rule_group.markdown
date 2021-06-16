---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_firewall_rule_group"
description: |-
  Provides a Route 53 Resolver DNS Firewall rule group resource.
---

# Resource: aws_route53_resolver_firewall_rule_group

Provides a Route 53 Resolver DNS Firewall rule group resource.

## Example Usage

```terraform
resource "aws_route53_resolver_firewall_rule_group" "example" {
  name = "example"
}
```

## Argument Reference

The following argument is supported:

* `name` - (Required) A name that lets you identify the rule group, to manage and use it.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN (Amazon Resource Name) of the rule group.
* `id` - The ID of the rule group.
* `owner_id` - The AWS account ID for the account that created the rule group. When a rule group is shared with your account, this is the account that has shared the rule group with you.
* `share_status` - Whether the rule group is shared with other AWS accounts, or was shared with the current account by another AWS account. Sharing is configured through AWS Resource Access Manager (AWS RAM). Valid values: `NOT_SHARED`, `SHARED_BY_ME`, `SHARED_WITH_ME`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

 Route 53 Resolver DNS Firewall rule groups can be imported using the Route 53 Resolver DNS Firewall rule group ID, e.g.

```
$ terraform import aws_route53_resolver_firewall_rule_group.example rslvr-frg-0123456789abcdef
```
