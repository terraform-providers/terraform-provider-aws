---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_regex_pattern_set"
description: |-
  Retrieves the summary of a WAFv2 Regex Pattern Set.
---

# Data Source: aws_wafv2_regex_pattern_set

Retrieves the summary of a WAFv2 Regex Pattern Set.

## Example Usage

```hcl
data "aws_wafv2_regex_pattern_set" "example" {
  name  = "some-regex-pattern-set"
  scope = "REGIONAL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAFv2 Regex Pattern Set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the entity.
* `description` - The description of the set that helps with identification.
* `id` - The unique identifier for the set.
* `regular_expression` - One or more blocks of regular expression patterns that AWS WAF is searching for. See [Regular Expression](#regular-expression) below for details.

### Regular Expression

Each `regular_expression` supports the following argument:

* `regex_string` - (Required) The string representing the regular expression, see the AWS WAF [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-regex-pattern-set-creating.html) for more information.
