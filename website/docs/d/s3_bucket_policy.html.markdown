---
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
sidebar_current: "docs-aws-datasource-s3-bucket-policy"
description: |-
  Provides details about a specific S3 bucket policy
---

# Data Source: aws_s3_bucket_policy

Provides details about a specific S3 bucket policy.

## Example Usage

### Merge bucket policy statements

```hcl
data "aws_s3_bucket" "example" {
  bucket = "bucket.test.com"
}

data "aws_s3_bucket_policy" "example" {
  bucket = "${data.aws_s3_bucket.example.id}"
}

data "aws_iam_policy_document" "example" {

  # use the bucket policy as the source
  source_json = "${data.aws_s3_bucket_policy.example.policy}"

  # overwrite any statements in the source policy which have matching statement IDs (sid)
  statement {
    sid = "ExampleStatement"
    actions   = ["s3:GetObject"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["1234567890"]
    }
  }
}

# update the bucket policy with the merged policy document
resource "aws_s3_bucket_policy" "example" {
  bucket = "${data.aws_s3_bucket.example.id}"
  policy = "${data.aws_iam_policy_document.example.json}"
}
```

## Argument Reference

The following arguments are supported:

- `bucket` - (Required) The name of the bucket

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `policy` - The policy for the bucket
