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

```terraform
resource "aws_glue_catalog_table" "aws_glue_catalog_table" {
  name          = "MyCatalogTable"
  database_name = "MyCatalogDatabase"
}
```

### Parquet Table for Athena

```terraform
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

The following arguments are required:

* `name` - (Required) Name of the table. For Hive compatibility, this must be entirely lowercase.
* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.

The follow arguments are optional:

* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `description` - (Optional) Description of the table.
* `owner` - (Optional) Owner of the table.
* `parameters` - (Optional) Properties associated with this table, as a list of key-value pairs.
* `partition_index` - (Optional) Configuration block for a maximum of 3 partition indexes. See [`partition_index`](#partition_index) below.
* `partition_keys` - (Optional) Configuration block of columns by which the table is partitioned. Only primitive types are supported as partition keys. See [`partition_keys`](#partition_keys) below.
* `retention` - (Optional) Retention time for this table.
* `storage_descriptor` - (Optional) Configuration block for information about the physical storage of this table. For more information, refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables. html#aws-glue-api-catalog-tables-StorageDescriptor). See [`storage_descriptor`](#storage_descriptor) below.
* `table_type` - (Optional) Type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
* `target_table` - (Optional) Configuration block of a target table for resource linking. See [`target_table`](#target_table) below.
* `view_expanded_text` - (Optional) If the table is a view, the expanded text of the view; otherwise null.
* `view_original_text` - (Optional) If the table is a view, the original text of the view; otherwise null.

### partition_index

* `index_name` - (Required) Name of the partition index.
* `keys` - (Required) Keys for the partition index.

### partition_keys

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Partition Key.
* `type` - (Optional) Datatype of data in the Partition Key.

### storage_descriptor

* `bucket_columns` - (Optional) List of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `columns` - (Optional) Configuration block for columns in the table. See [`columns`](#columns) below.
* `compressed` - (Optional) Whether the data in the table is compressed.
* `input_format` - (Optional) Input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `location` - (Optional) Physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `output_format` - (Optional) Output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `schema_reference` - (Optional) Object that references a schema stored in the AWS Glue Schema Registry. When creating a table, you can pass an empty list of columns for the schema, and instead use a schema reference. See [Schema Reference](#schema-reference) below.
* `ser_de_info` - (Optional) Configuration block for serialization and deserialization ("SerDe") information. See [`ser_de_info`](#ser_de_info) below.
* `skewed_info` - (Optional) Configuration block with information about values that appear very frequently in a column (skewed values). See [`skewed_info`](#skewed_info) below.
* `sort_columns` - (Optional) Configuration block for the sort order of each bucket in the table. See [`sort_columns`](#sort_columns) below.
* `stored_as_sub_directories` - (Optional) Whether the table data is stored in subdirectories.

#### columns

* `comment` - (Optional) Free-form text comment.
* `name` - (Required) Name of the Column.
* `parameters` - (Optional) Key-value pairs defining properties associated with the column.
* `type` - (Optional) Datatype of data in the Column.

#### schema_reference

* `schema_id` - (Optional) Configuration block that contains schema identity fields. Either this or the `schema_version_id` has to be provided. See [`schema_id`](#schema_id) below.
* `schema_version_id` - (Optional) Unique ID assigned to a version of the schema. Either this or the `schema_id` has to be provided.
* `schema_version_number` - (Required) Version number of the schema.

##### schema_id

* `registry_name` - (Optional) Name of the schema registry that contains the schema. Must be provided when `schema_name` is specified and conflicts with `schema_arn`.
* `schema_arn` - (Optional) ARN of the schema. One of `schema_arn` or `schema_name` has to be provided.
* `schema_name` - (Optional) Name of the schema. One of `schema_arn` or `schema_name` has to be provided.

#### ser_de_info

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) Map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is `org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe`.

#### sort_columns

* `column` - (Required) Name of the column.
* `sort_order` - (Required) Whether the column is sorted in ascending (`1`) or descending order (`0`).

#### skewed_info

* `skewed_column_names` - (Optional) List of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) List of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) Map of skewed values to the columns that contain them.

### target_table

* `catalog_id` - (Required) ID of the Data Catalog in which the table resides.
* `database_name` - (Required) Name of the catalog database that contains the target table.
* `name` - (Required) Name of the target table.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Glue Table.
* `id` - Catalog ID, Database name and of the name table.

## Import

Glue Tables can be imported with their catalog ID (usually AWS account ID), database name, and table name, e.g.

```
$ terraform import aws_glue_catalog_table.MyTable 123456789012:MyDatabase:MyTable
```
