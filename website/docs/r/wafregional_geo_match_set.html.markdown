---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_geo_match_set"
description: |-
  Provides a AWS WAF Regional Geo Match Set resource.
---

# Resource: aws_wafregional_geo_match_set

Provides a WAF Regional Geo Match Set Resource

## Example Usage

```hcl
resource "aws_wafregional_geo_match_set" "geo_match_set" {
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

* `name` - (Required) The name or description of the Geo Match Set.
* `geo_match_constraint` - (Optional) The Geo Match Constraint objects which contain the country that you want AWS WAF to search for.

## Nested Blocks

### `geo_match_constraint`

#### Arguments

* `type` - (Required) The type of geographical area you want AWS WAF to search for. Currently Country is the only valid value.
* `value` - (Required) The country that you want AWS WAF to search for.
  This is the two-letter country code, e.g. `US`, `CA`, `RU`, `CN`, etc.
  See [docs](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchConstraint.html) for all supported values.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional Geo Match Set.

## Import

WAF Regional Geo Match Set can be imported using the id, e.g.

```
$ terraform import aws_wafregional_geo_match_set.geo_match_set a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
