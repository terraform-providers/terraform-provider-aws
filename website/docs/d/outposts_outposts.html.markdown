---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outposts"
description: |-
  Provides details about multiple Outposts
---

# Data Source: aws_outposts_outposts

Provides details about multiple Outposts.

## Example Usage

```hcl
data "aws_outposts_outposts" "example" {
  site_id = data.aws_outposts_site.id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) Availability Zone name.
* `availability_zone_id` - (Optional) Availability Zone identifier.
* `site_id` - (Optional) Site identifier.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arns` - Set of Amazon Resource Names (ARNs).
* `id` - AWS Region.
* `ids` - Set of identifiers.
