---
layout: "aws"
page_title: "AWS: aws_glue_catalog_table"
sidebar_current: "docs-aws-resource-glue-catalog-table"
description: |-
  Provides a Glue Catalog Table.
---

# aws_glue_catalog_table

Provides a Glue Catalog Table Resource. You can refer to the [Glue Developer Guide](http://docs.aws.amazon.com/glue/latest/dg/populate-data-catalog.html) for a full explanation of the Glue Data Catalog functionality.

## Example Usage

```hcl
resource "aws_glue_catalog_table" "aws_glue_catalog_table" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.
* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.
* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `description` - (Optional) Description of the table.
* `owner` - (Optional) Owner of the table.
* `retention` - (Optional) Retention time for this table.
* `storage_descriptor` - (Optional) A [storage descriptor](#storage_descriptor) object containing information about the physical storage of this table. You can refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor) for a full explanation of this object.
* `partition_keys` - (Optional) A list of columns by which the table is partitioned. Only primitive types are supported as partition keys.
* `view_original_text` - (Optional) If the table is a view, the original text of the view; otherwise null.
* `view_expanded_text` - (Optional) If the table is a view, the expanded text of the view; otherwise null.
* `table_type` - (Optional) The type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.).
* `parameters` - (Optional) Properties associated with this table, as a list of key-value pairs.

##### storage_descriptor

* `columns` - (Optional) A list of the [Columns](#column) in the table.
* `location` - (Optional) The physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `input_format` - (Optional) The input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `output_format` - (Optional) The output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `compressed` - (Optional) True if the data in the table is compressed, or False if not.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `ser_de_info` - (Optional) [Serialization/deserialization (SerDe)](#ser_de_info) information.
* `bucket_columns` - (Optional) A list of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `sort_columns` - (Optional) A list of [Order](#sort_column) objects specifying the sort order of each bucket in the table.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `skewed_info` - (Optional) Information about values that appear very frequently in a column (skewed values).
* `stored_as_sub_directories` - (Optional) True if the table data is stored in subdirectories, or False if not.

##### column

* `name` - (Required) The name of the Column.
* `type` - (Optional) The datatype of data in the Column.
* `comment` - (Optional) Free-form text comment.

##### ser_de_info

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) A map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is: org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe.

##### sort_column

* `column` - (Required) The name of the column.
* `sort_order` - (Required) Indicates that the column is sorted in ascending order (== 1), or in descending order (==0).

##### skewed_info

* `skewed_column_names` - (Optional) A list of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) A list of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) A mapping of skewed values to the columns that contain them.

## Import

Glue Tables can be imported with their catalog ID (usually AWS account ID), database name, and table name, e.g.

```
$ terraform import aws_glue_catalog_table.MyTable 123456789012:MyDatabase:MyTable
```
