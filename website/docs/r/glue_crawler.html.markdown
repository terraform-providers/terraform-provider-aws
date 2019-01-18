---
layout: "aws"
page_title: "AWS: aws_glue_crawler"
sidebar_current: "docs-aws-resource-glue-crawler"
description: |-
  Manages a Glue Crawler
---

# aws_glue_crawler

Manages a Glue Crawler. More information can be found in the [AWS Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/add-crawler.html)

## Example Usage

### DynamoDB Target

```hcl
resource "aws_glue_crawler" "example" {
  database_name = "${aws_glue_catalog_database.example.name}"
  name          = "example"
  role          = "${aws_iam_role.example.arn}"

  dynamodb_target {
    path = "table-name"
  }
}
```

### JDBC Target

```hcl
resource "aws_glue_crawler" "example" {
  database_name = "${aws_glue_catalog_database.example.name}"
  name          = "example"
  role          = "${aws_iam_role.example.arn}"

  jdbc_target {
    connection_name = "${aws_glue_connection.example.name}"
    path            = "database-name/%"
  }
}
```

### S3 Target

```hcl
resource "aws_glue_crawler" "example" {
  database_name = "${aws_glue_catalog_database.example.name}"
  name          = "example"
  role          = "${aws_iam_role.example.arn}"

  s3_target {
    path = "s3://${aws_s3_bucket.example.bucket}"
  }
}
```

## Argument Reference

~> **NOTE:** At least one `jdbc_target` or `s3_target` must be specified.

The following arguments are supported:

* `database_name` (Required) Glue database where results are written.
* `name` (Required) Name of the crawler.
* `role` (Required) The IAM role friendly name (including path without leading slash), or ARN of an IAM role, used by the crawler to access other resources.
* `classifiers` (Optional) List of custom classifiers. By default, all AWS classifiers are included in a crawl, but these custom classifiers always override the default classifiers for a given classification.
* `configuration` (Optional) JSON string of configuration information.
* `description` (Optional) Description of the crawler.
* `dynamodb_target` (Optional) List of nested DynamoDB target arguments. See below.
* `jdbc_target` (Optional) List of nested JBDC target arguments. See below.
* `s3_target` (Optional) List nested Amazon S3 target arguments. See below.
* `schedule` (Optional) A cron expression used to specify the schedule. For more information, see [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html). For example, to run something every day at 12:15 UTC, you would specify: `cron(15 12 * * ? *)`.
* `schema_change_policy` (Optional) Policy for the crawler's update and deletion behavior.
* `table_prefix` (Optional) The table prefix used for catalog tables that are created.
* `security_configuration` (Optional) The name of Security Configuration to be used by the crawler

### dynamodb_target Argument Reference

* `path` - (Required) The name of the DynamoDB table to crawl.

### jdbc_target Argument Reference

* `connection_name` - (Required) The name of the connection to use to connect to the JDBC target.
* `path` - (Required) The path of the JDBC target.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.

### s3_target Argument Reference

* `path` - (Required) The path to the Amazon S3 target.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.

### schema_change_policy Argument Reference

* `delete_behavior` - (Optional) The deletion behavior when the crawler finds a deleted object. Valid values: `LOG`, `DELETE_FROM_DATABASE`, or `DEPRECATE_IN_DATABASE`. Defaults to `DEPRECATE_IN_DATABASE`.
* `update_behavior` - (Optional) The update behavior when the crawler finds a changed schema. Valid values: `LOG` or `UPDATE_IN_DATABASE`. Defaults to `UPDATE_IN_DATABASE`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Crawler name

## Import

Glue Crawlers can be imported using `name`, e.g.

```
$ terraform import aws_glue_crawler.MyJob MyJob
```
