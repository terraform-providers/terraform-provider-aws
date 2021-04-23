---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_permissions"
description: |-
    Get permissions for a principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3.
---

# Data Source: aws_lakeformation_permissions

Get permissions for a principal to access metadata in the Data Catalog and data organized in underlying data storage such as Amazon S3. Permissions are granted to a principal, in a Data Catalog, relative to a Lake Formation resource, which includes the Data Catalog, databases, and tables. For more information, see [Security and Access Control to Metadata and Data in Lake Formation](https://docs.aws.amazon.com/lake-formation/latest/dg/security-data-access.html).

~> **NOTE:** This data source deals with explicitly granted permissions. Lake Formation grants implicit permissions to data lake administrators, database creators, and table creators. For more information, see [Implicit Lake Formation Permissions](https://docs.aws.amazon.com/lake-formation/latest/dg/implicit-permissions.html).

## Example Usage

### Permissions For A Lake Formation S3 Resource

```hcl
data "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.workflow_role.arn

  data_location {
    arn = aws_lakeformation_resource.test.arn
  }
}
```

### Permissions For A Glue Catalog Database

```hcl
data "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.workflow_role.arn

  database {
    name       = aws_glue_catalog_database.test.name
    catalog_id = "110376042874"
  }
}
```

## Argument Reference

The following arguments are required:

* `principal` – (Required) Principal to be granted the permissions on the resource. Supported principals are IAM users or IAM roles.

One of the following is required:

* `catalog_resource` - Whether the permissions are to be granted for the Data Catalog. Defaults to `false`.
* `data_location` - Configuration block for a data location resource. Detailed below.
* `database` - Configuration block for a database resource. Detailed below.
* `table` - Configuration block for a table resource. Detailed below.
* `table_with_columns` - Configuration block for a table with columns resource. Detailed below.

The following arguments are optional:

* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID. The Data Catalog is the persistent metadata store. It contains database definitions, table definitions, and other control information to manage your Lake Formation environment.

### data_location

The following argument is required:

* `arn` – (Required) Amazon Resource Name (ARN) that uniquely identifies the data location resource.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog where the location is registered with Lake Formation. By default, it is the account ID of the caller.

### database

The following argument is required:

* `name` – (Required) Name of the database resource. Unique to the Data Catalog.

The following argument is optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.

### table

The following argument is required:

* `database_name` – (Required) Name of the database for the table. Unique to a Data Catalog.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `name` - (Optional) Name of the table. At least one of `name` or `wildcard` is required.
* `wildcard` - (Optional) Whether to use a wildcard representing every table under a database. At least one of `name` or `wildcard` is required. Defaults to `false`.

### table_with_columns

The following arguments are required:

* `database_name` – (Required) Name of the database for the table with columns resource. Unique to the Data Catalog.
* `name` – (Required) Name of the table resource.

The following arguments are optional:

* `catalog_id` - (Optional) Identifier for the Data Catalog. By default, it is the account ID of the caller.
* `column_names` - (Optional) List of column names for the table. At least one of `column_names` or `excluded_column_names` is required.
* `excluded_column_names` - (Optional) List of column names for the table to exclude. At least one of `column_names` or `excluded_column_names` is required.

## Attributes Reference

In addition to the above arguments, the following attribute is exported:

* `permissions` – List of permissions granted to the principal. For details on permissions, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).
* `permissions_with_grant_option` - Subset of `permissions` which the principal can pass.
