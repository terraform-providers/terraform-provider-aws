---
layout: "aws"
page_title: "AWS: waf_regex_pattern_set"
sidebar_current: "docs-aws-resource-waf-regex-pattern-set"
description: |-
  Provides a AWS WAF Regex Pattern Set resource.
---

# aws_waf_regex_pattern_set

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
