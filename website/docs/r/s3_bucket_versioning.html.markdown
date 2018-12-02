---
layout: "aws"
page_title: "AWS - aws_s3_bucket_versioning"
sidebar_current: "docs-aws-s3-bucket-versioning"
description: |-
  Provides a resource for controlling S3 bucket versioning
---

# aws_s3_bucket_versioning

Provides a resource for controlling versioning on an [S3 bucket][1].  Note that this
may conflict with the `versioning` block of an [aws_s3_bucket][1] if the settings are
not the same.

## Example Usage

```hcl
resource "aws_s3_bucket" "b" {
  bucket = "example-bucket"
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "versioning_example" {
  bucket  = "${aws_s3_bucket.b.id}"
  enabled = true
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the S3 bucket
* `enabled` - (Required) Whether to enable versioning on the bucket

## Deletion

Deleting this resource will disable versioning on the bucket.

[1]: /docs/providers/aws/r/s3_bucket.html
