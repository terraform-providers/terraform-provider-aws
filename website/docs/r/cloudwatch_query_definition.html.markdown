---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_query_definition"
description: |-
  Provides a CloudWatch Logs query definition resource.
---

# Resource: aws_cloudwatch_query_definition

Provides a CloudWatch Logs query definition resource.

## Example Usage

```terraform
resource "aws_cloudwatch_query_definition" "example" {
  name = "custom_query"

  log_group_names = [
    "/aws/logGroup1",
    "/aws/logGroup2"
  ]

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 25
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the query.
* `query_string` - (Required) The query to save. You can read more about CloudWatch Logs Query Syntax in the [documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CWL_QuerySyntax.html).
* `log_group_names` - (Optional) Specific log groups to use with the query.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `query_definition_id` - The query definition ID.

## Import

CloudWatch query definitions can be imported using the query definition ARN. The ARN can be found on the "Edit Query" page for the query in the AWS Console.

```
$ terraform import aws_cloudwatch_query_definition.example arn:aws:logs:us-west-2:123456789012:query-definition:269951d7-6f75-496d-9d7b-6b7a5486bdbd
```
