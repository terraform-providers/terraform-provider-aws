---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53_query_log"
description: |-
  Provides a Route53 query logging configuration resource.
---

# Resource: aws_route53_query_log

Provides a Route53 query logging configuration resource.

~> **NOTE:** There are restrictions on the configuration of query logging. Notably,
the CloudWatch log group must be in the `us-east-1` region,
a permissive CloudWatch log resource policy must be in place, and
the Route53 hosted zone must be public.
See [Configuring Logging for DNS Queries](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/query-logs.html?console_help=true#query-logs-configuring) for additional details.

## Example Usage

```hcl
# Example CloudWatch log group in us-east-1

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

resource "aws_cloudwatch_log_group" "aws_route53_example_com" {
  provider = aws.us-east-1

  name              = "/aws/route53/${aws_route53_zone.example_com.name}"
  retention_in_days = 30
}

# Example CloudWatch log resource policy to allow Route53 to write logs
# to any log group under /aws/route53/*

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
  provider = aws.us-east-1

  policy_document = data.aws_iam_policy_document.route53-query-logging-policy.json
  policy_name     = "route53-query-logging-policy"
}

# Example Route53 zone with query logging

resource "aws_route53_zone" "example_com" {
  name = "example.com"
}

resource "aws_route53_query_log" "example_com" {
  depends_on = [aws_cloudwatch_log_resource_policy.route53-query-logging-policy]

  cloudwatch_log_group_arn = aws_cloudwatch_log_group.aws_route53_example_com.arn
  zone_id                  = aws_route53_zone.example_com.zone_id
}
```

## Argument Reference

The following arguments are supported:

* `cloudwatch_log_group_arn` - (Required) CloudWatch log group ARN to send query logs.
* `zone_id` - (Required) Route53 hosted zone ID to enable query logs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The query logging configuration ID

## Import

Route53 query logging configurations can be imported using their ID, e.g.

```
$ terraform import aws_route53_query_log.example_com xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
