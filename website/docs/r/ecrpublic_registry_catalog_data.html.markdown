---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecrpublic_registry_catalog_data"
description: |-
  Provides the ability to manage catalog data for an ECR Public Registry
---

# Resource: aws_ecrpublic_registry_catalog_data

Provides the ability to manage catalog data for an ECR Public Registry

## Example Usage

```hcl
resource "aws_ecrpublic_registry_catalog_data" "foo" {
  display_name = "A display name for the registry"
}
```

## Argument Reference

The following arguments are supported:

* `display_name` - (Required) The display name for a public registry. This appears on the Amazon ECR Public Gallery. Only accounts that have the verified account badge can have a registry display name.


## Import

ECR Public Registry Catalog Data can be imported using the `name`, e.g.

```
$ terraform import aws_ecrpublic_registry_catalog_data.example example
```
