---
subcategory: "S3 Control"
layout: "aws"
page_title: "AWS: aws_s3control_bucket_lifecycle_configuration"
description: |-
  Manages an S3 Control Bucket Lifecycle Configuration.
---

# Resource: aws_s3control_bucket_lifecycle_configuration

Provides a resource to manage an S3 Control Bucket Lifecycle Configuration.

~> **NOTE:** Each S3 Control Bucket can only have one Lifecycle Configuration. Using multiple of this resource against the same S3 Control Bucket will result in perpetual differences each Terraform run.

-> This functionality is for managing [S3 on Outposts](https://docs.aws.amazon.com/AmazonS3/latest/dev/S3onOutposts.html). To manage S3 Bucket Lifecycle Configurations in an AWS Partition, see the [`aws_s3_bucket` resource](/docs/providers/aws/r/s3_bucket.html).

## Example Usage

```hcl
resource "aws_s3control_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3control_bucket.example.arn

  rule {
    expiration {
      days = 365
    }

    filter {
      prefix = "logs/"
    }

    id = "logs"
  }

  rule {
    expiration {
      days = 7
    }

    filter {
      prefix = "temp/"
    }

    id = "temp"
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Amazon Resource Name (ARN) of the bucket.
* `rule` - (Required) Configuration block(s) containing lifecycle rules for the bucket.
    * `abort_incomplete_multipart_upload` - (Optional) Configuration block containing settings for abort incomplete multipart upload.
        * `days_after_initiation` - (Required) Number of days after which Amazon S3 aborts an incomplete multipart upload.
    * `expiration` - (Optional) Configuration block containing settings for expiration of objects.
        * `date` - (Optional) Date the object is to be deleted. Should be in `YYYY-MM-DD` date format, e.g. `2020-09-30`.
        * `days` - (Optional) Number of days before the object is to be deleted.
        * `expired_object_delete_marker` - (Optional) Enable to remove a delete marker with no noncurrent versions. Cannot be specified with `date` or `days`.
    * `filter` - (Optional) Configuration block containing settings for filtering.
        * `prefix` - (Optional) Object prefix for rule filtering.
        * `tags` - (Optional) Key-value map of object tags for rule filtering.
    * `id` - (Required) Unique identifier for the rule.
    * `status` - (Optional) Status of the rule. Valid values: `Enabled` and `Disabled`. Defaults to `Enabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the bucket.

## Import

S3 Control Bucket Lifecycle Configurations can be imported using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_s3control_bucket_lifecycle_configuration.example arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-12345678/bucket/example
```
