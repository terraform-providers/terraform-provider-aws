---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_partition"
description: |-
  Provides a Glue Partition.
---

# Resource: aws_glue_partition

Provides a Glue Partition Resource.

## Example Usage

```hcl
resource "aws_glue_partition" "example" {
  database_name = "some-database"
  table_name    = "some-table"
  values        = ["some-value"]
}
```

## Argument Reference

The following arguments are supported:

* `database_name` - (Required) Name of the metadata database where the table metadata resides. For Hive compatibility, this must be all lowercase.
* `partition_values` - (Required) The values that define the partition.
* `catalog_id` - (Optional) ID of the Glue Catalog and database to create the table in. If omitted, this defaults to the AWS Account ID plus the database name.
* `storage_descriptor` - (Optional) A [storage descriptor](#storage_descriptor) object containing information about the physical storage of this table. You can refer to the [Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-api-catalog-tables.html#aws-glue-api-catalog-tables-StorageDescriptor) for a full explanation of this object.
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

* `id` - partition id.
* `creation_time` - The time at which the partition was created.
* `last_analyzed_time` - The last time at which column statistics were computed for this partition.
* `last_accessed_time` - The last time at which the partition was accessed.

## Import

Glue Partitions can be imported with their catalog ID (usually AWS account ID), database name, table name and partition values e.g.

```
$ terraform import aws_glue_partition.part 123456789012:MyDatabase:MyTable:val1#val2
```
