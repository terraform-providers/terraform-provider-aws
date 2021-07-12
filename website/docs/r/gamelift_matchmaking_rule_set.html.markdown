---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_matchmaking_rule_set"
description: |-
  Provides a Gamelift Matchmaking Rule Set resource.
---

# Resource: aws_gamelift_matchmaking_rule_set

Provides an Gamelift Matchmaking Rule Set resource.

## Example Usage

```terraform
resource "aws_gamelift_matchmaking_rule_set" "test" {
  name = "example-rule-set"

  rule_set_body = <<RULE_SET_BODY
{
  "name": "rule-set",
  "ruleLanguageVersion": "1.0",
  "teams": [{
    "name": "alpha",
    "minPlayers": 1,
    "maxPlayers": 5
  }]"
}
RULE_SET_BODY
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the rule set.
* `rule_set_body` - (Required) A collection of matchmaking rules, formatted as a JSON string. Comments are not allowed in JSON, but most elements support a description field.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Rule Set ARN.

## Import

Gamelift Match Making Rule Sets can be imported by their `name`, e.g.

```
$ terraform import aws_gamelift_matchmaking_rule_set.example example
```
