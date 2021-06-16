---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_function"
description: |-
  Provides a CloudFront Function resource. With CloudFront Functions in Amazon CloudFront, you can write lightweight functions in JavaScript for high-scale, latency-sensitive CDN customizations.
---

# Resource: aws_cloudfront_function

Provides a CloudFront Function resource. With CloudFront Functions in Amazon CloudFront, you can write lightweight functions in JavaScript for high-scale, latency-sensitive CDN customizations.

See [CloudFront Functions](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html)

~> **NOTE:** You cannot delete a function if it’s associated with a cache behavior. First, update your distributions to remove the function association from all cache behaviors, then delete the function.

## Example Usage

### Basic Example

```terraform
resource "aws_cloudfront_function" "test" {
  name    = "test"
  runtime = "cloudfront-js-1.0"
  comment = "my function"
  publish = true
  code    = file("${path.module}/function.js")
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Unique name for your CloudFront Function.
* `code` - (Required) Source code of the function
* `runtime` - (Required) Identifier of the function's runtime. Currently only `cloudfront-js-1.0` is valid.

The following arguments are optional:

* `comment` - (Optional) Comment.
* `publish` - (Optional) Whether to publish creation/change as Live CloudFront Function Version. Defaults to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifying your CloudFront Function.
* `etag` - ETag hash of the function
* `status` - Status of the function. Can be `UNPUBLISHED`, `UNASSOCIATED` or `ASSOCIATED`.

## Import

CloudFront Functions can be imported using the `name`, e.g.

```
$ terraform import aws_cloudfront_function.test my_test_function
```
