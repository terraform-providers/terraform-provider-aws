---
layout: "aws"
page_title: "AWS: aws_appsync_datasource"
sidebar_current: "docs-aws-resource-appsync-datasource"
description: |-
  Provides an AppSync DataSource.
---

# aws_appsync_datasource

Provides an AppSync DataSource.

## Example Usage

```hcl
resource "aws_dynamodb_table" "example" {
  name           = "example"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserId"

  attribute {
    name = "UserId"
    type = "S"
  }
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "example"
  role = "${aws_iam_role.example.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_dynamodb_table.example.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_appsync_graphql_api" "example" {
  authentication_type = "API_KEY"
  name                = "tf_appsync_example"
}

resource "aws_appsync_datasource" "example" {
  api_id           = "${aws_appsync_graphql_api.example.id}"
  name             = "tf_appsync_example"
  service_role_arn = "${aws_iam_role.example.arn}"
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    table_name = "${aws_dynamodb_table.example.name}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API ID for the GraphQL API for the DataSource.
* `name` - (Required) A user-supplied name for the DataSource.
* `type` - (Required) The type of the DataSource. Valid values: `AWS_LAMBDA`, `AMAZON_DYNAMODB`, `AMAZON_ELASTICSEARCH`, `HTTP`, `NONE`.
* `description` - (Optional) A description of the DataSource.
* `service_role_arn` - (Optional) The IAM service role ARN for the data source.
* `dynamodb_config` - (Optional) DynamoDB settings. See [below](#dynamodb_config)
* `elasticsearch_config` - (Optional) Amazon Elasticsearch settings. See [below](#elasticsearch_config)
* `http_config` - (Optional) HTTP settings. See [below](#http_config)
* `lambda_config` - (Optional) AWS Lambda settings. See [below](#lambda_config)

### dynamodb_config

The following arguments are supported:

* `table_name` - (Required) Name of the DynamoDB table.
* `region` - (Optional) AWS region of the DynamoDB table. Defaults to current region.
* `use_caller_credentials` - (Optional) Set to `true` to use Amazon Cognito credentials with this data source.

### elasticsearch_config

The following arguments are supported:

* `endpoint` - (Required) HTTP endpoint of the Elasticsearch domain.
* `region` - (Optional) AWS region of Elasticsearch domain. Defaults to current region.

### http_config

The following arguments are supported:

* `endpoint` - (Required) HTTP URL.

### lambda_config

The following arguments are supported:

* `function_arn` - (Required) The ARN for the Lambda function.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN

## Import

`aws_appsync_datasource` can be imported with their `api_id`, a hyphen, and `name`, e.g.

```
$ terraform import aws_appsync_datasource.example abcdef123456-example
```
