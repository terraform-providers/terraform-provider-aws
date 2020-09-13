---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_metric"
description: |-
  Provides a S3 bucket metrics configuration resource.
---

# Resource: aws_s3_bucket_metric

Provides a S3 bucket [metrics configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/metrics-configurations.html) resource.

## Example Usage

### Add metrics configuration for entire S3 bucket

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_metric" "example-entire-bucket" {
  bucket = aws_s3_bucket.example.bucket
  name   = "EntireBucket"
}
```

### Add metrics configuration with S3 bucket object filter

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_metric" "example-filtered" {
  bucket = aws_s3_bucket.example.bucket
  name   = "ImportantBlueDocuments"

  filter {
    prefix = "documents/"

    tags = {
      priority = "high"
      class    = "blue"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to put metric configuration.
* `name` - (Required) Unique identifier of the metrics configuration for the bucket.
* `filter` - (Optional) [Object filtering](http://docs.aws.amazon.com/AmazonS3/latest/dev/metrics-configurations.html#metrics-configurations-filter) that accepts a prefix, tags, or a logical AND of prefix and tags (documented below).

The `filter` metric configuration supports the following:

~> **NOTE**: At least one of `prefix` or `tags` is required when specifying a `filter`

* `prefix` - (Optional) Object prefix for filtering (singular).
* `tags` - (Optional) Object tags for filtering (up to 10).

## Import

S3 bucket metric configurations can be imported using `bucket:metric`, e.g.

```
$ terraform import aws_s3_bucket_metric.my-bucket-entire-bucket my-bucket:EntireBucket
```
