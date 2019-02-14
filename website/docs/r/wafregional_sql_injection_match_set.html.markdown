---
layout: "aws"
page_title: "AWS: wafregional_sql_injection_match_set"
sidebar_current: "docs-aws-resource-wafregional-sql-injection-match-set"
description: |-
  Provides a AWS WAF Regional SqlInjectionMatchSet resource for use with ALB.
---

# aws_wafregional_sql_injection_match_set

Provides a WAF Regional SQL Injection Match Set Resource for use with Application Load Balancer.

## Example Usage

```hcl
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "tf-sql_injection_match_set"

  sql_injection_match_tuple {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the SizeConstraintSet.
* `sql_injection_match_tuple` - (Optional) The parts of web requests that you want AWS WAF to inspect for malicious SQL code and, if you want AWS WAF to inspect a header, the name of the header.

### Nested fields

### `sql_injection_match_tuple`

* `field_to_match` - (Required) Specifies where in a web request to look for snippets of malicious SQL code.
* `text_transformation` - (Required) Text transformations used to eliminate unusual formatting that attackers use in web requests in an effort to bypass AWS WAF.
  If you specify a transformation, AWS WAF performs the transformation on `field_to_match` before inspecting a request for a match.
  e.g. `CMD_LINE`, `HTML_ENTITY_DECODE` or `NONE`.
  See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_SqlInjectionMatchTuple.html#WAF-Type-regional_SqlInjectionMatchTuple-TextTransformation)
  for all supported values.

### `field_to_match`

* `data` - (Optional) When `type` is `HEADER`, enter the name of the header that you want to search, e.g. `User-Agent` or `Referer`.
  If `type` is any other value, omit this field.
* `type` - (Required) The part of the web request that you want AWS WAF to search for a specified string.
  e.g. `HEADER`, `METHOD` or `BODY`.
  See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_FieldToMatch.html)
  for all supported values.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF SqlInjectionMatchSet.
