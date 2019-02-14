---
layout: "aws"
page_title: "AWS: wafregional_regex_pattern_set"
sidebar_current: "docs-aws-resource-wafregional-regex-pattern-set"
description: |-
  Provides a AWS WAF Regional Regex Pattern Set resource.
---

# aws_wafregional_regex_pattern_set

Provides a WAF Regional Regex Pattern Set Resource

## Example Usage

```hcl
resource "aws_wafregional_regex_pattern_set" "example" {
  name                  = "example"
  regex_pattern_strings = ["one", "two"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the Regex Pattern Set.
* `regex_pattern_strings` - (Optional) A list of regular expression (regex) patterns that you want AWS WAF to search for, such as `B[a@]dB[o0]t`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional Regex Pattern Set.
