---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_bucket"
description: |-
  Manages an S3 Control Bucket.
---

# Resource: aws_s3control_bucket

Provides a resource to manage an S3 Control Bucket.

-> This functionality is for managing [S3 on Outposts](https://docs.aws.amazon.com/AmazonS3/latest/dev/S3onOutposts.html). To manage S3 Buckets in an AWS Partition, see the [`aws_s3_bucket` resource](/docs/providers/aws/r/s3_bucket.html).

## Example Usage

```hcl
resource "aws_s3control_bucket" "example" {
  bucket     = "example"
  outpost_id = data.aws_outposts_outpost.example.id
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Name of the bucket.
* `outpost_id` - (Required) Identifier of the Outpost to contain this bucket.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the bucket.
* `creation_date` - UTC creation date in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `id` - Amazon Resource Name (ARN) of the bucket.
* `public_access_block_enabled` - Boolean whether Public Access Block is enabled.

## Import

S3 Control Buckets can be imported using Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_s3control_bucket.example arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-12345678/bucket/example
```