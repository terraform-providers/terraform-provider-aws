---
layout: "aws"
page_title: "AWS: aws_custom_resource"
sidebar_current: "docs-aws-resource-custom-resource"
description: |-
  Provides an adapter for AWS CloudFormation Custom Resources.
  This allows you to perform "Create", "Update" and "Delete"
  operations on compliant Lambda functions or SNS topics.
---

# Resource: aws_custom_resource

Provides an adapter for AWS CloudFormation Custom Resources. This allows you to perform "Create", "Update" and "Delete" operations on compliant Lambda functions or SNS topics.

## Example Usage

```hcl
resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"
  force_detach_policies = true
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "lambda_logging" {
  name = "lambda_logging"
  path = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_lambda_function" "lambda_function" {
  filename      = "custom_resource.zip"
  function_name = "custom_resource_lambda"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs10.x"
}

resource "aws_custom_resource" "custom_resource" {
  service_token = aws_lambda_function.lambda_function.arn
  resource_type = "CustomTest"
  resource_properties = {
    a = "1"
    b = "2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_token` - (Required) The service token (an Amazon SNS topic or AWS Lambda function Amazon Resource Name) that is obtained from the custom resource provider to access the service. 
* `resource_type` - (Required) The developer-chosen resource type of the custom resource
* `resource_properties` - (Optional) This field contains the contents of the Properties object sent by the Terraform. Its contents are defined by the custom resource provider.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A random id that is unique to this resource.
* `old_resource_properties` - Used only for Update requests. Contains the resource properties that were declared previous to the update request.
* `data` - The custom resource provider-defined name-value pairs sent with the response.

## Import

Custom Resources can be imported using the `id`, e.g.

```
$ terraform import aws_custom_resource.custom_resource 5577006791947779410
```
