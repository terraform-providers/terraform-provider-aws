---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_catalog_table"
description: |-
  Provides a Glue Catalog Table.
---

# Resource: aws_glue_catalog_table

Provides a Glue Catalog Table Resource. You can refer to the [Glue Developer Guide](http://docs.aws.amazon.com/glue/latest/dg/populate-data-catalog.html) for a full explanation of the Glue Data Catalog functionality.

## Example Usage

### Basic Table

```hcl
resource "aws_glue_catalog_table" "aws_glue_catalog_table" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"
}
```

### Parquet Table for Athena

```hcl
resource "aws_glue_catalog_table" "aws_glue_catalog_table" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"

  table_type = "EXTERNAL_TABLE"

  parameters = {
    EXTERNAL              = "TRUE"
    "parquet.compression" = "SNAPPY"
  }

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"

      parameters = {
        "serialization.format" = 1
      }
    }

    columns {
      name = "my_string"
      type = "string"
    }

    columns {
      name = "my_double"
      type = "double"
    }

    columns {
      name    = "my_date"
      type    = "date"
      comment = ""
    }

    columns {
      name    = "my_bigint"
      type    = "bigint"
      comment = ""
    }

    columns {
      name    = "my_struct"
      type    = "struct<my_nested_string:string>"
      comment = ""
    }
  }
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
* `table_type` - (Optional) The type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
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

##### sort_columns

* `column` - (Required) The name of the column.
* `sort_order` - (Required) Indicates that the column is sorted in ascending order (== 1), or in descending order (==0).

##### skewed_info

* `skewed_column_names` - (Optional) A list of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) A list of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) A map of skewed values to the columns that contain them.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Catalog ID, Database name and of the name table.
* `arn` - The ARN of the Glue Table.


## Import

Glue Tables can be imported with their catalog ID (usually AWS account ID), database name, and table name, e.g.

```
$ terraform import aws_glue_catalog_table.MyTable 123456789012:MyDatabase:MyTable
```
