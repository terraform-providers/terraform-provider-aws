---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_bucket_policy"
description: |-
  Manages an S3 Control Bucket Policy.
---

# Resource: aws_s3control_bucket_policy

Provides a resource to manage an S3 Control Bucket Policy.

-> This functionality is for managing [S3 on Outposts](https://docs.aws.amazon.com/AmazonS3/latest/dev/S3onOutposts.html). To manage S3 Bucket Policies in an AWS Partition, see the [`aws_s3_bucket_policy` resource](/docs/providers/aws/r/s3_bucket_policy.html).

## Example Usage

```hcl
resource "aws_s3control_bucket_policy" "example" {
  bucket = aws_s3control_bucket.example.arn
  policy = jsonencode({
    Id = "testBucketPolicy"
    Statement = [
      {
        Action = "s3-outposts:PutBucketLifecycleConfiguration"
        Effect = "Deny"
        Principal = {
          AWS = "*"
        }
        Resource = aws_s3control_bucket.example.arn
        Sid      = "statement1"
      }
    ]
    Version = "2012-10-17"
  })
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Amazon Resource Name (ARN) of the bucket.
* `policy` - (Required) JSON string of the resource policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the bucket.

## Import

S3 Control Bucket Policies can be imported using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_s3control_bucket_policy.example arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-12345678/bucket/example
```
