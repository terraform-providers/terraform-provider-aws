---
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
sidebar_current: "docs-aws-resource-s3-bucket-policy"
description: |-
  Attaches a policy to an S3 bucket resource.
---

# aws_s3_bucket_policy

Attaches a policy to an S3 bucket resource.

## Example Usage

### Basic Usage

```hcl
resource "aws_s3_bucket" "b" {
  bucket = "my_tf_test_bucket"
}

resource "aws_s3_bucket_policy" "b" {
  bucket = "${aws_s3_bucket.b.id}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "MYBUCKETPOLICY",
  "Statement": [
    {
      "Sid": "IPAllow",
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:*",
      "Resource": "arn:aws:s3:::my_tf_test_bucket/*",
      "Condition": {
         "IpAddress": {"aws:SourceIp": "8.8.8.8/32"}
      }
    }
  ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to which to apply the policy.
* `policy` - (Required) The text of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](/docs/providers/aws/guides/iam-policy-documents.html).

## Import

S3 bucket policies can be imported using the bucket name, e.g.

```
$ terraform import aws_s3_bucket_policy.example my-bucket-name
```
