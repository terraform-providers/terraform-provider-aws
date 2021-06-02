---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_function"
description: |-
  Provides a CloudFront Function data source.
---

# aws_cloudfront_function

Provides information about a CloudFront Function.

## Example Usage

```terraform
variable "function_name" {
  type = string
}

data "aws_cloudfront_function" "existing" {
  name = var.function_name
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the CloudFront function.
* `stage` - (Required) The function’s stage, either `DEVELOPMENT` or `LIVE`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifying your CloudFront Function.
* `code` - Source code of the function
* `comment` - Comment.
* `etag` - ETag hash of the function
* `last_modified_time` - When this resource was last modified.
* `runtime` - Identifier of the function's runtime.
* `status` - Status of the function. Can be `UNPUBLISHED`, `UNASSOCIATED` or `ASSOCIATED`.
