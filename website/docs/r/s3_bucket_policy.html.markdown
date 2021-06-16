---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
description: |-
  Attaches a policy to an S3 bucket resource.
---

# Resource: aws_s3_bucket_policy

Attaches a policy to an S3 bucket resource.

## Example Usage

### Basic Usage

```terraform
resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
}

resource "aws_s3_bucket_policy" "b" {
  bucket = aws_s3_bucket.b.id

  # Terraform's "jsonencode" function converts a
  # Terraform expression's result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "MYBUCKETPOLICY"
    Statement = [
      {
        Sid       = "IPAllow"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.b.arn,
          "${aws_s3_bucket.b.arn}/*",
        ]
        Condition = {
          IpAddress = {
            "aws:SourceIp" = "8.8.8.8/32"
          }
        }
      },
    ]
  })
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to which to apply the policy.
* `policy` - (Required) The text of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy). Note: Bucket policies are limited to 20 KB in size.

## Attributes Reference

No additional attributes are exported.

## Import

S3 bucket policies can be imported using the bucket name, e.g.

```
$ terraform import aws_s3_bucket_policy.example my-bucket-name
```
