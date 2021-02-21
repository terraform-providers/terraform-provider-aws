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
* `storage_descriptor` - (Optional) A [storage descriptor](#storage-descriptor) object containing information about the physical storage of this table. You can refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor) for a full explanation of this object.
* `partition_keys` - (Optional) A list of columns by which the table is partitioned. Only primitive types are supported as partition keys. see [Partition Keys](#partition-keys) below.
* `view_original_text` - (Optional) If the table is a view, the original text of the view; otherwise null.
* `view_expanded_text` - (Optional) If the table is a view, the expanded text of the view; otherwise null.
* `table_type` - (Optional) The type of this table (EXTERNAL_TABLE, VIRTUAL_VIEW, etc.). While optional, some Athena DDL queries such as `ALTER TABLE` and `SHOW CREATE TABLE` will fail if this argument is empty.
* `parameters` - (Optional) Properties associated with this table, as a list of key-value pairs.
* `partition_index` - (Optional) A list of partition indexes. see [Partition Index](#partition-index) below.

### Partition Index

* `index_name` - (Required) The name of the partition index.
* `keys` - (Required) The keys for the partition index.

### Partition Keys

* `name` - (Required) The name of the Partition Key.
* `type` - (Optional) The datatype of data in the Partition Key.
* `comment` - (Optional) Free-form text comment.

### Storage Descriptor

* `columns` - (Optional) A list of the [Columns](#column) in the table.
* `location` - (Optional) The physical location of the table. By default this takes the form of the warehouse location, followed by the database location in the warehouse, followed by the table name.
* `input_format` - (Optional) The input format: SequenceFileInputFormat (binary), or TextInputFormat, or a custom format.
* `output_format` - (Optional) The output format: SequenceFileOutputFormat (binary), or IgnoreKeyTextOutputFormat, or a custom format.
* `compressed` - (Optional) True if the data in the table is compressed, or False if not.
* `number_of_buckets` - (Optional) Must be specified if the table contains any dimension columns.
* `ser_de_info` - (Optional) [Serialization/deserialization (SerDe)](#ser-de-info) information.
* `bucket_columns` - (Optional) A list of reducer grouping columns, clustering columns, and bucketing columns in the table.
* `sort_columns` - (Optional) A list of [Order](#sort-column) objects specifying the sort order of each bucket in the table.
* `parameters` - (Optional) User-supplied properties in key-value form.
* `skewed_info` - (Optional) Information about values that appear very frequently in a column (skewed values).
* `stored_as_sub_directories` - (Optional) True if the table data is stored in subdirectories, or False if not.
* `schema_reference` - (Optional) An object that references a schema stored in the AWS Glue Schema Registry. When creating a table, you can pass an empty list of columns for the schema, and instead use a schema reference. See [Schema Reference](#schema-reference) below.

##### Column

* `name` - (Required) The name of the Column.
* `type` - (Optional) The datatype of data in the Column.
* `comment` - (Optional) Free-form text comment.
* `parameters` - (Optional) These key-value pairs define properties associated with the column.

##### Ser De Info

* `name` - (Optional) Name of the SerDe.
* `parameters` - (Optional) A map of initialization parameters for the SerDe, in key-value form.
* `serialization_library` - (Optional) Usually the class that implements the SerDe. An example is: org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe.

##### Sort Columns

* `column` - (Required) The name of the column.
* `sort_order` - (Required) Indicates that the column is sorted in ascending order (== 1), or in descending order (==0).

##### Skewed Info

* `skewed_column_names` - (Optional) A list of names of columns that contain skewed values.
* `skewed_column_value_location_maps` - (Optional) A list of values that appear so frequently as to be considered skewed.
* `skewed_column_values` - (Optional) A map of skewed values to the columns that contain them.

##### Schema Reference

* `schema_id` - (Optional) A structure that contains schema identity fields. Either this or the `schema_version_id` has to be provided. See [Schema ID](#schema-id) below.
* `schema_version_id` - (Optional) The unique ID assigned to a version of the schema. Either this or the `schema_id` has to be provided.
* `schema_version_number` - (Required) The version number of the schema.

###### Schema ID

* `schema_arn` - (Optional) The Amazon Resource Name (ARN) of the schema. One of `schema_arn` or `schema_name` has to be provided.
* `schema_name` - (Optional) The name of the schema. One of `schema_arn` or `schema_name` has to be provided.
* `registry_name` - (Optional) The name of the schema registry that contains the schema. Must be provided when `schema_name` is specified and conflicts with `schema_arn`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Catalog ID, Database name and of the name table.
* `arn` - The ARN of the Glue Table.


## Import

Glue Tables can be imported with their catalog ID (usually AWS account ID), database name, and table name, e.g.

```
$ terraform import aws_glue_catalog_table.MyTable 123456789012:MyDatabase:MyTable
```
