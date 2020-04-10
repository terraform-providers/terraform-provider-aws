---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_waf_regex_pattern_set"
description: |-
  Provides a AWS WAF Regex Pattern Set resource.
---

# Resource: aws_waf_regex_pattern_set

Provides a WAF Regex Pattern Set Resource

## Example Usage

```hcl
resource "aws_waf_regex_pattern_set" "example" {
  name                  = "tf_waf_regex_pattern_set"
  regex_pattern_strings = ["one", "two"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the Regex Pattern Set.
* `regex_pattern_strings` - (Optional) A list of regular expression (regex) patterns that you want AWS WAF to search for, such as `B[a@]dB[o0]t`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regex Pattern Set.
* `arn` - Amazon Resource Name (ARN)

## Import

AWS WAF Regex Pattern Set can be imported using their ID, e.g.

```
$ terraform import aws_waf_regex_pattern_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
