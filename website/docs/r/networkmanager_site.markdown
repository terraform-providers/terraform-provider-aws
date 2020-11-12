---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_site"
description: |-
  Creates a site in a global network.
---

# Resource: aws_networkmanager_site

Creates a site in a global network.

## Example Usage

```hcl
resource "aws_networkmanager_global_network" "example" {
}

resource "aws_networkmanager_site" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}
```

## Argument Reference

The following arguments are supported:

* `global_network_id` - (Required) The ID of the Global Network to create the site in.
* `description` - (Optional) Description of the Site.
* `location` - (Optional) The site location as documented below.
* `tags` - (Optional) Key-value tags for the Site.

The `location` object supports the following:

* `address` - (Optional) Address of the location.
* `latitude` - (Optional) Latitude of the location.
* `longitude` - (Optional) Longitude of the location.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Site Amazon Resource Name (ARN)

## Import

`aws_networkmanager_site` can be imported using the site ARN, e.g.

```
$ terraform import aws_networkmanager_site.example arn:aws:networkmanager::123456789012:site/global-network-0d47f6t230mz46dy4/site-444555aaabbb11223
```
