---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_cors_configuration"
description: |-
  Provides a S3 bucket CORS configuration resource.
---

# Resource: aws_s3_bucket_cors_configuration

Provides a S3 bucket CORS configuration resource.

## Example Usage

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
  acl    = "public-read"
}

resource "aws_s3_bucket_cors_configuration" "cors_policy" {
    bucket = aws_s3_bucket.example.bucket

    cors_rule {
        allowed_headers = ["*"]
        allowed_methods = ["PUT", "POST"]
        allowed_origins = ["https://www.cors-configuration-test-create.com"]
        expose_headers  = ["x-amz-server-side-encryption", "ETag"]
        max_age_seconds = 3000
    }
}

```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to apply the CORS configuration to.

* `cors_rule` - (Required) A rule of [Cross-Origin Resource Sharing](https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) (documented below).

The `CORS` object supports the following:

* `allowed_headers` (Optional) Specifies which headers are allowed.
* `allowed_methods` (Required) Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.
* `allowed_origins` (Required) Specifies which origins are allowed.
* `expose_headers` (Optional) Specifies expose header in the response.
* `max_age_seconds` (Optional) Specifies time in seconds that browser can cache the response for a preflight request.


