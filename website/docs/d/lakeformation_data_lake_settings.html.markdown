---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_lake_settings"
description: |-
    Get data lake administrators and default database and table permissions
---

# Data Source: aws_lakeformation_data_lake_settings

Get Lake Formation principals designated as data lake administrators and lists of principal permission entries for default create database and default create table permissions.

## Example Usage

```hcl
data "aws_lakeformation_data_lake_settings" "example" {
  catalog_id = "14916253649"
}
```

## Argument Reference

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID.

## Attributes Reference

In addition to arguments above, the following attributes are exported.

* `admins` – List of ARNs of AWS Lake Formation principals (IAM users or roles).
* `create_database_default_permissions` - Up to three configuration blocks of principal permissions for default create database permissions. Detailed below.
* `create_table_default_permissions` - Up to three configuration blocks of principal permissions for default create table permissions. Detailed below.
* `trusted_resource_owners` – List of the resource-owning account IDs that the caller's account can use to share their user access details (user ARNs).

### create_database_default_permissions

* `permissions` - List of permissions granted to the principal.
* `principal` - Principal who is granted permissions.

### create_table_default_permissions

* `permissions` - List of permissions granted to the principal.
* `principal` - Principal who is granted permissions.
