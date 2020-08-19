---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_source"
description: |-
  Manages a Resource QuickSight Data Source.
---

# Resource: aws_quicksight_data_source

Resource for managing QuickSight Data Source

## Example Usage

```hcl
resource "aws_quicksight_data_source" "default" {
  data_source_id = "abcdefg"
  name           = "My Cool Data in S3"

  parameters {
    s3 {
      manifest_file_location {
        bucket = "my.bucket"
        key = "path/to/manifest.json"
      }
    }
  }
}
```

## Argument Reference

The QuickSight data source argument layout is a complex structure composed
of several sub-resources - these resources are laid out below.

### Top-Level Arguments

 * `name` - (Required) A name for the data source.

 * `data_source_id` - (Required) An identifier for the data source.

 * `aws_account_id` - (Optional) The ID for the AWS account that the data source is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.

 * `parameters` - (Required) The [parameters](#parameters-arguments) used to connect to this data source (exactly one).

#### Parameters Arguments

To specify data source connection parameters, exactly one of the following sub-objects must be provided.

 * `amazon_elasticsearch` - [Parameters](#amazon-elasticsearch-arguments) for connecting to Amazon Elasticsearch.

 * `athena` - [Parameters](#athena-arguments) for connecting to Athena.

 * `aurora` - [Parameters](#aurora-arguments) for connecting to Athena.

 * `aurora_postgresql` - [Parameters](#aurora-postgresql-arguments) for connecting to Aurora Postgresql.

 * `aws_iot_analytics` - [Parameters](#aws-iot-analytics-arguments) for connecting to AWS IOT Analytics.

 * `jira` - [Parameters](#jira-arguments) for connecting to Jira.

 * `maria_db` - [Parameters](#mariadb-arguments) for connecting to MariaDB.

 * `mysql` - [Parameters](#mysql-arguments) for connecting to MySQL.

 * `postgresql` - [Parameters](#postgresql-arguments) for connecting to Postgresql.

 * `presto` - [Parameters](#presto-arguments) for connecting to Presto.

 * `redshift` - [Parameters](#redshift-arguments) for connecting to Redshift.

 * `s3` - [Parameters](#s3-arguments) for connecting to S3.

 * `service_now` - [Parameters](#servicenow-arguments) for connecting to ServiceNow.

 * `snowflake` - [Parameters](#snowflake-arguments) for connecting to Snowflake.

 * `spark` - [Parameters](#spark-arguments) for connecting to SPARK.

 * `sql_server` - [Parameters](#sqlserver-arguments) for connecting to SqlServer.

 * `teradata` - [Parameters](#teradata-arguments) for connecting to Teradata.

 * `twitter` - [Parameters](#twitter-arguments) for connecting to Twitter.

#### Amazon Elasticsearch Arguments

 * `domain` - (Required) The domain to which to connect.

#### Athena Arguments

 * `work_group` - (Optional) The work-group to which to connect.

#### Aurora Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### Aurora Postgresql Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### AWS IOT Analytics Postgresql Arguments

 * `data_set_name` - (Required) The name of the data set to which to connect.

#### Jira Arguments

 * `site_base_url` - (Required) The base URL of the Jira instance's site to which to connect.

#### MariaDB Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### MySQL Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### Postgresql Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### Presto Arguments

 * `catalog` - (Required) The catalog to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The port to which to connect.

#### Redshift Arguments

 * `cluster_id` - (Optional) The ID of the cluster to which to connect.

 * `database` - (Required) The database to which to connect.

 * `host` - (Optional) The host to which to connect.

 * `port` - (Optional) The port to which to connect.

#### S3 Arguments

 * `manifest_file_location` - (Required) An [object containing the S3 location](#manifest-file-location-arguments) of the S3 manifest file.

##### Manifest File Location Arguments

 * `bucket` - (Required) The name of the bucket that contains the manifest file.

 * `key` - (Required) The key of the manifest file within the bucket.


#### ServiceNow Arguments

 * `site_base_url` - (Required) The base URL of the Jira instance's site to which to connect.

#### Snowflake Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `warehouse` - (Required) The warehouse to which to connect.

#### SPARK Arguments

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The warehouse to which to connect.

#### SqlServer Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The warehouse to which to connect.

#### Teradata Arguments

 * `database` - (Required) The database to which to connect.

 * `host` - (Required) The host to which to connect.

 * `port` - (Required) The warehouse to which to connect.

#### Twitter Arguments

 * `max_rows` - (Required) The maximum number of rows to query.

 * `query` - (Required) The Twitter query to retrieve the data.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the data source

* `type` - A key indicating which data source type was inferred from the passed `parameters`

## Import

A QuickSight data source can be imported using the AWS account ID, and data source ID name separated by `/`.

```
$ terraform import aws_quicksight_data_source.example 123456789123/my-data-source-id
```
