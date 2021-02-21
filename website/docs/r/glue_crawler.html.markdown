---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_crawler"
description: |-
  Manages a Glue Crawler
---

# Resource: aws_glue_crawler

Manages a Glue Crawler. More information can be found in the [AWS Glue Developer Guide](https://docs.aws.amazon.com/glue/latest/dg/add-crawler.html)

## Example Usage

### DynamoDB Target Example

```hcl
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  dynamodb_target {
    path = "table-name"
  }
}
```

### JDBC Target Example

```hcl
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  jdbc_target {
    connection_name = aws_glue_connection.example.name
    path            = "database-name/%"
  }
}
```

### S3 Target Example

```hcl
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  s3_target {
    path = "s3://${aws_s3_bucket.example.bucket}"
  }
}
```


### Catalog Target Example

```hcl
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  catalog_target {
    database_name = aws_glue_catalog_database.example.name
    tables        = [aws_glue_catalog_table.example.name]
  }

  schema_change_policy {
    delete_behavior = "LOG"
  }

  configuration = <<EOF
{
  "Version":1.0,
  "Grouping": {
    "TableGroupingPolicy": "CombineCompatibleSchemas"
  }
}
EOF
}
```

### MongoDB Target Example

```hcl
resource "aws_glue_crawler" "example" {
  database_name = aws_glue_catalog_database.example.name
  name          = "example"
  role          = aws_iam_role.example.arn

  mongodb_target {
    connection_name = aws_glue_connection.example.name
    path            = "database-name/%"
  }
}
```

### Configuration Settings Example

```hcl
resource "aws_glue_crawler" "events_crawler" {
  database_name = aws_glue_catalog_database.glue_database.name
  schedule      = "cron(0 1 * * ? *)"
  name          = "events_crawler_${var.environment_name}"
  role          = aws_iam_role.glue_role.arn
  tags          = var.tags

  configuration = jsonencode(
    {
      Grouping = {
        TableGroupingPolicy = "CombineCompatibleSchemas"
      }
      CrawlerOutput = {
        Partitions = { AddOrUpdateBehavior = "InheritFromTable" }
      }
      Version = 1
    }
  )

  s3_target {
    path = "s3://${aws_s3_bucket.data_lake_bucket.bucket}"
  }
}
```

## Argument Reference

~> **NOTE:** Must specify at least one of `dynamodb_target`, `jdbc_target`, `s3_target` or `catalog_target`.

The following arguments are supported:

* `database_name` (Required) Glue database where results are written.
* `name` (Required) Name of the crawler.
* `role` (Required) The IAM role friendly name (including path without leading slash), or ARN of an IAM role, used by the crawler to access other resources.
* `classifiers` (Optional) List of custom classifiers. By default, all AWS classifiers are included in a crawl, but these custom classifiers always override the default classifiers for a given classification.
* `configuration` (Optional) JSON string of configuration information. For more details see [Setting Crawler Configuration Options](https://docs.aws.amazon.com/glue/latest/dg/crawler-configuration.html).
* `description` (Optional) Description of the crawler.
* `dynamodb_target` (Optional) List of nested DynamoDB target arguments. See [Dynamodb Target](#dynamodb-target) below.
* `jdbc_target` (Optional) List of nested JBDC target arguments. See [JDBC Target](#jdbc-target) below.
* `s3_target` (Optional) List nested Amazon S3 target arguments. See [S3 Target](#s3-target) below.
* `mongodb_target` (Optional) List nested MongoDB target arguments. See [MongoDB Target](#mongodb-target) below.
* `schedule` (Optional) A cron expression used to specify the schedule. For more information, see [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html). For example, to run something every day at 12:15 UTC, you would specify: `cron(15 12 * * ? *)`.
* `schema_change_policy` (Optional) Policy for the crawler's update and deletion behavior. See [Schema Change Policy](#schema-change-policy) below.
* `lineage_configuration` (Optional) Specifies data lineage configuration settings for the crawler. See [Lineage Configuration](#lineage-configuration) below.
* `recrawl_policy` (Optional)  A policy that specifies whether to crawl the entire dataset again, or to crawl only folders that were added since the last crawler run.. See [Recrawl Policy](#recrawl-policy) below.
* `security_configuration` (Optional) The name of Security Configuration to be used by the crawler
* `table_prefix` (Optional) The table prefix used for catalog tables that are created.
* `tags` - (Optional) Key-value map of resource tags

### Dynamodb Target

* `path` - (Required) The name of the DynamoDB table to crawl.
* `scan_all` - (Optional) Indicates whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table.  defaults to `true`.
* `scan_rate` - (Optional) The percentage of the configured read capacity units to use by the AWS Glue crawler. The valid values are null or a value between 0.1 to 1.5.

### JDBC Target

* `connection_name` - (Required) The name of the connection to use to connect to the JDBC target.
* `path` - (Required) The path of the JDBC target.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.

### S3 Target

* `path` - (Required) The path to the Amazon S3 target.
* `connection_name` - (Optional) The name of a connection which allows crawler to access data in S3 within a VPC.
* `exclusions` - (Optional) A list of glob patterns used to exclude from the crawl.

### Catalog Target

* `database_name` - (Required) The name of the Glue database to be synchronized.
* `tables` - (Required) A list of catalog tables to be synchronized.

~> **Note:** `deletion_behavior` of catalog target doesn't support `DEPRECATE_IN_DATABASE`.

-> **Note:** `configuration` for catalog target crawlers will have `{ ... "Grouping": { "TableGroupingPolicy": "CombineCompatibleSchemas"} }` by default.


### MongoDB Target

* `connection_name` - (Required) The name of the connection to use to connect to the Amazon DocumentDB or MongoDB target.
* `path` - (Required) The path of the Amazon DocumentDB or MongoDB target (database/collection).
* `scan_all` - (Optional) Indicates whether to scan all the records, or to sample rows from the table. Scanning all the records can take a long time when the table is not a high throughput table. Default value is `true`.

### Schema Change Policy

* `delete_behavior` - (Optional) The deletion behavior when the crawler finds a deleted object. Valid values: `LOG`, `DELETE_FROM_DATABASE`, or `DEPRECATE_IN_DATABASE`. Defaults to `DEPRECATE_IN_DATABASE`.
* `update_behavior` - (Optional) The update behavior when the crawler finds a changed schema. Valid values: `LOG` or `UPDATE_IN_DATABASE`. Defaults to `UPDATE_IN_DATABASE`.

### Lineage Configuration

* `crawler_lineage_settings` - (Optional) Specifies whether data lineage is enabled for the crawler. Valid values are: `ENABLE` and `DISABLE`. Default value is `Disable`.

### Recrawl Policy

* `recrawl_behavior` - (Optional) Specifies whether to crawl the entire dataset again or to crawl only folders that were added since the last crawler run. Valid Values are: `CRAWL_EVERYTHING` and `CRAWL_NEW_FOLDERS_ONLY`. Default value is `CRAWL_EVERYTHING`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Crawler name
* `arn` - The ARN of the crawler

## Import

Glue Crawlers can be imported using `name`, e.g.

```
$ terraform import aws_glue_crawler.MyJob MyJob
```
