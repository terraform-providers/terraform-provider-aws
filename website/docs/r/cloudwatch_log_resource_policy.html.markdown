---
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_resource_policy"
sidebar_current: "docs-aws-resource-cloudwatch-log-resource-policy"
description: |-
  Provides a resource to manage a CloudWatch log resource policy
---

# aws_cloudwatch_log_resource_policy

Provides a resource to manage a CloudWatch log resource policy.

## Example Usage

### Elasticsearch Log Publishing

```hcl
data "aws_iam_policy_document" "elasticsearch-log-publishing-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:PutLogEventsBatch",
    ]

    resources = ["arn:aws:logs:*"]

    principals {
      identifiers = ["es.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "elasticsearch-log-publishing-policy" {
  policy_document = "${data.aws_iam_policy_document.elasticsearch-log-publishing-policy.json}"
  policy_name     = "elasticsearch-log-publishing-policy"
}
```

### Route53 Query Logging

```hcl
data "aws_iam_policy_document" "route53-query-logging-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "route53-query-logging-policy" {
  policy_document = "${data.aws_iam_policy_document.route53-query-logging-policy.json}"
  policy_name     = "route53-query-logging-policy"
}
```

## Argument Reference

The following arguments are supported:

* `policy_document` - (Required) Details of the resource policy, including the identity of the principal that is enabled to put logs to this account. This is formatted as a JSON string. Maximum length of 5120 characters.
* `policy_name` - (Required) Name of the resource policy.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the CloudWatch log resource policy

## Import

CloudWatch log resource policies can be imported using the policy name, e.g.

```
$ terraform import aws_cloudwatch_log_resource_policy.MyPolicy MyPolicy
```
