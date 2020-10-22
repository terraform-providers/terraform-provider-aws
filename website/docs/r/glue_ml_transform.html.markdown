---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_ml_transform"
description: |-
  Provides a Glue ML Transform resource.
---

# Resource: aws_glue_ml_transform

Provides a Glue ML Transform resource.

## Example Usage

```hcl
resource "aws_glue_ml_transform" "test" {
  name     = "example"
  role_arn = aws_iam_role.test.arn

  input_record_tables {
    database_name = aws_glue_catalog_table.test.database_name
    table_name    = aws_glue_catalog_table.test.name
  }

  parameters {
    transform_type = "FIND_MATCHES"

    find_matches_parameters {
      primary_key_column_name = "my_column_1"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_glue_catalog_database" "test" {
  name = "example"
}

resource "aws_glue_catalog_table" "test" {
  name               = "example"
  database_name      = aws_glue_catalog_database.test.name
  owner              = "my_owner"
  retention          = 1
  table_type         = "VIRTUAL_VIEW"
  view_expanded_text = "view_expanded_text_1"
  view_original_text = "view_original_text_1"

  storage_descriptor {
    bucket_columns            = ["bucket_column_1"]
    compressed                = false
    input_format              = "SequenceFileInputFormat"
    location                  = "my_location"
    number_of_buckets         = 1
    output_format             = "SequenceFileInputFormat"
    stored_as_sub_directories = false

    parameters = {
      param1 = "param1_val"
    }

    columns {
      name    = "my_column_1"
      type    = "int"
      comment = "my_column1_comment"
    }

    columns {
      name    = "my_column_2"
      type    = "string"
      comment = "my_column2_comment"
    }

    ser_de_info {
      name = "ser_de_name"

      parameters = {
        param1 = "param_val_1"
      }

      serialization_library = "org.apache.hadoop.hive.serde2.columnar.ColumnarSerDe"
    }

    sort_columns {
      column     = "my_column_1"
      sort_order = 1
    }

    skewed_info {
      skewed_column_names = [
        "my_column_1",
      ]

      skewed_column_value_location_maps = {
        my_column_1 = "my_column_1_val_loc_map"
      }

      skewed_column_values = [
        "skewed_val_1",
      ]
    }
  }

  partition_keys {
    name    = "my_column_1"
    type    = "int"
    comment = "my_column_1_comment"
  }

  partition_keys {
    name    = "my_column_2"
    type    = "string"
    comment = "my_column_2_comment"
  }

  parameters = {
    param1 = "param1_val"
  }
}
```


## Argument Reference

The following arguments are supported:

* `name` – (Required) The name you assign to this ML Transform. It must be unique in your account.
* `input_record_tables` - (Required)  A list of AWS Glue table definitions used by the transform. see [Input Record Tables](#input_record_tables).
* `parameters` - (Required) The algorithmic parameters that are specific to the transform type used. Conditionally dependent on the transform type. see [Parameters](#parameters).
* `role_arn` – (Required) The ARN of the IAM role associated with this ML Transform.
* `description` – (Optional) Description of the ML Transform.
* `glue_version` - (Optional) The version of glue to use, for example "1.0". For information about available versions, see the [AWS Glue Release Notes](https://docs.aws.amazon.com/glue/latest/dg/release-notes.html).
* `max_capacity` – (Optional) The number of AWS Glue data processing units (DPUs) that are allocated to task runs for this transform. You can allocate from `2` to `100` DPUs; the default is `10`. `max_capacity` is a mutually exclusive option with `number_of_workers` and `worker_type`.
* `max_retries` – (Optional) The maximum number of times to retry this ML Transform if it fails.
* `tags` - (Optional) Key-value map of resource tags
* `timeout` – (Optional) The ML Transform timeout in minutes. The default is 2880 minutes (48 hours).
* `worker_type` - (Optional) The type of predefined worker that is allocated when an ML Transform runs. Accepts a value of `Standard`, `G.1X`, or `G.2X`. Required with `number_of_workers`.
* `number_of_workers` - (Optional) The number of workers of a defined `worker_type` that are allocated when an ML Transform runs. Required with `worker_type`.

### input_record_tables

* `database_name` - (Required) A database name in the AWS Glue Data Catalog.
* `table_name` - (Required) A table name in the AWS Glue Data Catalog.
* `catalog_id` - (Optional) A unique identifier for the AWS Glue Data Catalog.
* `connection_name`- (Optional) The name of the connection to the AWS Glue Data Catalog.

### parameters

* `transform_type` - (Required) The type of machine learning transform. For information about the types of machine learning transforms, see [Creating Machine Learning Transforms](http://docs.aws.amazon.com/glue/latest/dg/add-job-machine-learning-transform.html).
* `find_matches_parameters` - (Required) The parameters for the find matches algorithm. see [Find Matches Parameters](#find_matches_parameters).

#### find_matches_parameters

* `accuracy_cost_trade_off` - (Optional) The value that is selected when tuning your transform for a balance between accuracy and cost.
* `enforce_provided_labels` - (Optional) The value to switch on or off to force the output to match the provided labels from users.
* `precision_recall_trade_off` - (Optional) The value selected when tuning your transform for a balance between precision and recall.
* `primary_key_column_name` - (Optional) The name of a column that uniquely identifies rows in the source table.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Glue ML Transform.
* `id` - Glue ML Transform ID.
* `label_count` - The number of labels available for this transform.
* `schema` - The object that represents the schema that this transform accepts. see [Schema](#schema).

### schema

* `name` - The name of the column.
* `data_type` - The type of data in the column.

## Import

Glue ML Transforms can be imported using `id`, e.g.

```
$ terraform import aws_glue_ml_transform.example tfm-c2cafbe83b1c575f49eaca9939220e2fcd58e2d5
```
