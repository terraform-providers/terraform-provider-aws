---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS:  aws_networkmanager_site"
description: |-
  Retrieve information about a site.
---

# Data Source:  aws_networkmanager_site

Retrieve information about a site.

## Example Usage

```hcl
data " aws_networkmanager_site" "example" {
  id = var.global_network_id
}
```

## Argument Reference

* `id` - (Optional) The id of the specific site to retrieve.

* `global_network_id` - (Required) The ID of the Global Network of the site to retrieve.

* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the desired site.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the site.
* `description` - The description of the site.
* `location` - The site location as documented below.
* `tags` - Key-value tags for the Site.

The `location` object supports the following:

* `address` - Address of the location.
* `latitude` - Latitude of the location.
* `longitude` - Longitude of the location.
