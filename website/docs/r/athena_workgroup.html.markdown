---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_workgroup"
description: |-
  Manages an Athena Workgroup.
---

# Resource: aws_athena_workgroup

Provides an Athena Workgroup.

## Example Usage

```hcl
resource "aws_athena_workgroup" "example" {
  name = "example"

  configuration {
    enforce_workgroup_configuration    = true
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.example.bucket}/output/"

      encryption_configuration {
        encryption_option = "SSE_KMS"
        kms_key_arn       = aws_kms_key.example.arn
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the workgroup.
* `configuration` - (Optional) Configuration block with various settings for the workgroup. Documented below.
* `description` - (Optional) Description of the workgroup.
* `state` - (Optional) State of the workgroup. Valid values are `DISABLED` or `ENABLED`. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags for the workgroup.
* `force_destroy` - (Optional) The option to delete the workgroup and its contents even if the workgroup contains any named queries.

### configuration Argument Reference

The `configuration` configuration block supports the following arguments:

* `bytes_scanned_cutoff_per_query` - (Optional) Integer for the upper data usage limit (cutoff) for the amount of bytes a single query in a workgroup is allowed to scan. Must be at least `10485760`.
* `enforce_workgroup_configuration` - (Optional) Boolean whether the settings for the workgroup override client-side settings. For more information, see [Workgroup Settings Override Client-Side Settings](https://docs.aws.amazon.com/athena/latest/ug/workgroups-settings-override.html). Defaults to `true`.
* `publish_cloudwatch_metrics_enabled` - (Optional) Boolean whether Amazon CloudWatch metrics are enabled for the workgroup. Defaults to `true`.
* `result_configuration` - (Optional) Configuration block with result settings. Documented below.

#### result_configuration Argument Reference

The `result_configuration` configuration block within the `configuration` supports the following arguments:

* `encryption_configuration` - (Optional) Configuration block with encryption settings. Documented below.
* `output_location` - (Optional) The location in Amazon S3 where your query results are stored, such as `s3://path/to/query/bucket/`. For more information, see [Queries and Query Result Files](https://docs.aws.amazon.com/athena/latest/ug/querying.html).

##### encryption_configuration Argument Reference

The `encryption_configuration` configuration block within the `result_configuration` of the `configuration` supports the following arguments:

* `encryption_option` - (Required) Indicates whether Amazon S3 server-side encryption with Amazon S3-managed keys (`SSE_S3`), server-side encryption with KMS-managed keys (`SSE_KMS`), or client-side encryption with KMS-managed keys (`CSE_KMS`) is used. If a query runs in a workgroup and the workgroup overrides client-side settings, then the workgroup's setting for encryption is used. It specifies whether query results must be encrypted, for all queries that run in this workgroup.
* `kms_key_arn` - (Optional) For `SSE_KMS` and `CSE_KMS`, this is the KMS key Amazon Resource Name (ARN).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the workgroup
* `id` - The workgroup name

## Import

Athena Workgroups can be imported using their name, e.g.

```
$ terraform import aws_athena_workgroup.example example
```
