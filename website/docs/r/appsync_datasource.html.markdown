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
  name = "example"
  read_capacity  = 1
  write_capacity = 1
  hash_key = "UserId"
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
  name = "tf_appsync_example"
}

resource "aws_appsync_datasource" "example" {
  api_id = "${aws_appsync_graphql_api.example.id}"
  name = "tf_appsync_example"
  type = "AMAZON_DYNAMODB"
  dynamodb_config {
    region = "us-west-2"
    table_name = "${aws_dynamodb_table.example.name}"
  }
  service_role_arn = "${aws_iam_role.example.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `api_id` - (Required) The API ID for the GraphQL API for the DataSource.
* `name` - (Required) A user-supplied name for the DataSource.
* `type` - (Required) The type of the DataSource. Valid values: `AWS_LAMBDA`, `AMAZON_DYNAMODB` and `AMAZON_ELASTICSEARCH`
* `description` - (Optional) A description of the DataSource.
* `service_role_arn` - (Optional) The IAM service role ARN for the data source.
* `dynamodb_config` - (Optional) DynamoDB settings. See [below](#dynamodb_config)
* `elasticsearch_config` - (Optional) Amazon Elasticsearch settings. See [below](#elasticsearch_config)
* `lambda_config` - (Optional) AWS Lambda settings. See [below](#lambda_config)

### dynamodb_config

The following arguments are supported:

* `region` - (Required) The AWS region.
* `table_name` - (Required) The table name.
* `use_caller_credentials` - (Optional) Set to TRUE to use Amazon Cognito credentials with this data source.

### elasticsearch_config

The following arguments are supported:

* `region` - (Required) The AWS region.
* `endpoint` - (Required) The endpoint.

### lambda_config

The following arguments are supported:

* `function_arn` - (Required) The ARN for the Lambda function.

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN
