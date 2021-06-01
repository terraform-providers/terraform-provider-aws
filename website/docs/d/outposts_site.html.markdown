---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_site"
description: |-
  Provides details about an Outposts Site
---

# Data Source: aws_outposts_site

Provides details about an Outposts Site.

## Example Usage

```hcl
data "aws_outposts_site" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) Identifier of the Site.
* `name` - (Optional) Name of the Site.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - AWS Account identifier.
* `description` - Description.
