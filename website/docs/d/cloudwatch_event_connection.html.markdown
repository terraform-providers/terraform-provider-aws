---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_connection"
description: |-
  Provides an EventBridge connection data source.
---

# Data source: aws_cloudfront_distribution

Use this data source to retrieve information about a EventBridge connection.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.


## Example Usage

```terraform
data "aws_cloudwatch_event_connection" "test" {
  name = "test"
}
```

## Argument Reference

* `name` - The name of the connection.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the connection.

* `arn` - The ARN (Amazon Resource Name) for the connection.

* `secret_arn` - The ARN (Amazon Resource Name) for the secret created from the authorization parameters specified for the connection.

* `authorization_type` - The type of authorization to use to connect. One of `API_KEY`,`BASIC`,`OAUTH_CLIENT_CREDENTIALS`.
