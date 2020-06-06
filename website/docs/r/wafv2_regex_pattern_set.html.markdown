---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_regex_pattern_set"
description: |-
  Provides an AWS WAFv2 Regex Pattern Set resource.
---

# Resource: aws_wafv2_regex_pattern_set

Provides an AWS WAFv2 Regex Pattern Set Resource

## Example Usage

```hcl
resource "aws_wafv2_regex_pattern_set" "example" {
  name        = "example"
  description = "Example regex pattern set"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name of the regular expression pattern set.
* `description` - (Optional) A friendly description of the regular expression pattern set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.
* `regular_expression` - (Optional) One or more blocks of regular expression patterns that you want AWS WAF to search for, such as `B[a@]dB[o0]t`. See [Regular Expression](#regular-expression) below for details.
* `tags` - (Optional) An array of key:value pairs to associate with the resource.

### Regular Expression

* `regex_string` - (Required) The string representing the regular expression, see the AWS WAF [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-regex-pattern-set-creating.html) for more information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for the set.
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster.

## Import

WAFv2 Regex Pattern Sets can be imported using `ID/name/scope` e.g.

```
$ terraform import aws_wafv2_regex_pattern_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
