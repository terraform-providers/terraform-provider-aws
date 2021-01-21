---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_lake_settings"
description: |-
  Manages data lake administrators and default database and table permissions
---

# Resource: aws_lakeformation_data_lake_settings

Manages Lake Formation principals designated as data lake administrators and lists of principal permission entries for default create database and default create table permissions.

~> **NOTE:** Lake Formation introduces fine-grained access control for data in your data lake. Part of the changes include the `IAMAllowedPrincipals` principal in order to make Lake Formation backwards compatible with existing IAM and Glue permissions. For more information, see [Changing the Default Security Settings for Your Data Lake](https://docs.aws.amazon.com/lake-formation/latest/dg/change-settings.html) and [Upgrading AWS Glue Data Permissions to the AWS Lake Formation Model](https://docs.aws.amazon.com/lake-formation/latest/dg/upgrade-glue-lake-formation.html).

## Example Usage

### Data Lake Admins

```hcl
resource "aws_lakeformation_data_lake_settings" "example" {
  admins = [aws_iam_user.test.arn, aws_iam_role.test.arn]
}
```

### Create Default Permissions

```hcl
resource "aws_lakeformation_data_lake_settings" "example" {
  admins = [aws_iam_user.test.arn, aws_iam_role.test.arn]

  create_database_default_permissions {
    permissions = ["SELECT", "ALTER", "DROP"]
    principal   = aws_iam_user.test.arn
  }

  create_table_default_permissions {
    permissions = ["ALL"]
    principal   = aws_iam_role.test.arn
  }
}
```

## Argument Reference

The following arguments are optional:

* `admins` – (Optional) List of ARNs of AWS Lake Formation principals (IAM users or roles).
* `catalog_id` – (Optional) Identifier for the Data Catalog. By default, the account ID.
* `create_database_default_permissions` - (Optional) Up to three configuration blocks of principal permissions for default create database permissions. Detailed below.
* `create_table_default_permissions` - (Optional) Up to three configuration blocks of principal permissions for default create table permissions. Detailed below.
* `trusted_resource_owners` – (Optional) List of the resource-owning account IDs that the caller's account can use to share their user access details (user ARNs).

~> **NOTE:** Although optional, not including `admins`, `create_database_default_permissions`, `create_table_default_permissions`, and/or `trusted_resource_owners` results in the setting being cleared.

### create_database_default_permissions

The following arguments are optional:

* `permissions` - (Optional) List of permissions that are granted to the principal. Valid values may include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, `DESCRIBE`, and `CREATE_TABLE`. For more details, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).
* `principal` - (Optional) Principal who is granted permissions. To enforce metadata and underlying data access control only by IAM on new databases and tables set `principal` to `IAM_ALLOWED_PRINCIPALS` and `permissions` to `["ALL"]`.

### create_table_default_permissions

The following arguments are optional:

* `permissions` - (Optional) List of permissions that are granted to the principal. Valid values may include `ALL`, `SELECT`, `ALTER`, `DROP`, `DELETE`, `INSERT`, and `DESCRIBE`. For more details, see [Lake Formation Permissions Reference](https://docs.aws.amazon.com/lake-formation/latest/dg/lf-permissions-reference.html).
* `principal` - (Optional) Principal who is granted permissions. To enforce metadata and underlying data access control only by IAM on new databases and tables set `principal` to `IAM_ALLOWED_PRINCIPALS` and `permissions` to `["ALL"]`.

## Attributes Reference

In addition to all arguments above, no attributes are exported.
