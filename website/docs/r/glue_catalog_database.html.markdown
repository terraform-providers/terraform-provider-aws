---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_database"
description: |-
  Provides a Glue Catalog Database.
---

# Resource: aws_glue_catalog_database

Provides a Glue Catalog Database Resource. You can refer to the [Glue Developer Guide](http://docs.aws.amazon.com/glue/latest/dg/populate-data-catalog.html) for a full explanation of the Glue Data Catalog functionality

## Example Usage

```terraform
resource "aws_glue_catalog_database" "aws_glue_catalog_database" {
  name = "MyCatalogDatabase"
}
```

## Argument Reference

The following arguments are supported:

* `catalog_id` - (Optional) ID of the Glue Catalog to create the database in. If omitted, this defaults to the AWS Account ID.
* `description` - (Optional) Description of the database.
* `location_uri` - (Optional) Location of the database (for example, an HDFS path).
* `name` - (Required) Name of the database. The acceptable characters are lowercase letters, numbers, and the underscore character.
* `parameters` - (Optional) List of key-value pairs that define parameters and properties of the database.
* `target_database` - (Optional) Configuration block for a target database for resource linking. See [`target_database`](#target_database) below.

### target_database

* `catalog_id` - (Required) ID of the Data Catalog in which the database resides.
* `database_name` - (Required) Name of the catalog database.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Glue Catalog Database.
* `id` - Catalog ID and name of the database

## Import

Glue Catalog Databases can be imported using the `catalog_id:name`. If you have not set a Catalog ID specify the AWS Account ID that the database is in, e.g.

```
$ terraform import aws_glue_catalog_database.database 123456789012:my_database
```
