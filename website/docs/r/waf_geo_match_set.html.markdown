---
layout: "aws"
page_title: "AWS: waf_geo_match_set"
sidebar_current: "docs-aws-resource-waf-geo-match-set"
description: |-
  Provides a AWS WAF GeoMatchSet resource.
---

# aws_waf_geo_match_set

Provides a WAF Geo Match Set Resource

## Example Usage

```hcl
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = "geo_match_set"

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }

  geo_match_constraint {
    type  = "Country"
    value = "CA"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the GeoMatchSet.
* `geo_match_constraint` - (Optional) The GeoMatchConstraint objects which contain the country that you want AWS WAF to search for.

## Nested Blocks

### `geo_match_constraint`

#### Arguments

* `type` - (Required) The type of geographical area you want AWS WAF to search for. Currently Country is the only valid value.
* `value` - (Required) The country that you want AWS WAF to search for.
  This is the two-letter country code, e.g. `US`, `CA`, `RU`, `CN`, etc.
  See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchConstraint.html) for all supported values.

## Remarks

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the WAF GeoMatchSet.
