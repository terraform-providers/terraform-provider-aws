---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_ownership_controls"
description: |-
  Manages S3 Bucket Ownership Controls.
---

# Resource: aws_s3_bucket_ownership_controls

Provides a resource to manage S3 Bucket Ownership Controls. For more information, see the [S3 Developer Guide](https://docs.aws.amazon.com/AmazonS3/latest/dev/about-object-ownership.html).

~> **NOTE:** This AWS functionality is in Preview and may change before General Availability release. Backwards compatibility is not guaranteed between Terraform AWS Provider releases.

## Example Usage

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_ownership_controls" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) The name of the bucket that you want to associate this access point with.
* `rule` - (Required) Configuration block(s) with Ownership Controls rules. Detailed below.

### rule Configuration Block

The following arguments are required:

* `object_ownership` - (Optional) Object ownership. Valid values: `BucketOwnerPreferred` or `ObjectWriter`
    * `BucketOwnerPreferred` - Objects uploaded to the bucket change ownership to the bucket owner if the objects are uploaded with the `bucket-owner-full-control` canned ACL.
    * `ObjectWriter` - The uploading account will own the object if the object is uploaded with the `bucket-owner-full-control` canned ACL.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - S3 Bucket name.

## Import

S3 Bucket Ownership Controls can be imported using S3 Bucket name, e.g.

```
$ terraform import aws_s3_bucket_ownership_controls.example my-bucket
```
