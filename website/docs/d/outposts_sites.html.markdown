---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_sites"
description: |-
  Provides details about multiple Outposts Sites.
---

# Data Source: aws_outposts_sites

Provides details about multiple Outposts Sites.

## Example Usage

```hcl
data "aws_outposts_sites" "all" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

The following attributes are exported:

* `id` - AWS Region.
* `ids` - Set of Outposts Site identifiers.
