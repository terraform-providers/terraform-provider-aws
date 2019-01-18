---
layout: "aws"
page_title: "AWS: aws_s3_bucket_public_access_block"
sidebar_current: "docs-aws-resource-s3-bucket-public-access-block"
description: |-
  Manages S3 bucket-level Public Access Block Configuration
---

# aws_s3_bucket_public_access_block

Manages S3 bucket-level Public Access Block configuration. For more information about these settings, see the [AWS S3 Block Public Access documentation](https://docs.aws.amazon.com/AmazonS3/latest/dev/access-control-block-public-access.html).

## Example Usage

```hcl
resource "aws_s3_bucket" "example" {
    bucket = "example"
}

resource "aws_s3_bucket_public_access_block" "example" {
  bucket = "${aws_s3_bucket.bucket.id}"

  block_public_acls   = true
  block_public_policy = true
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) S3 Bucket to which this Public Access Block configuration should be applied.
* `block_public_acls` - (Optional) Whether Amazon S3 should block public ACLs for this bucket. Defaults to `false`. Enabling this setting does not affect existing policies or ACLs. When set to `true` causes the following behavior:
  * PUT Bucket acl and PUT Object acl calls will fail if the specified ACL allows public access.
  * PUT Object calls will fail if the request includes an object ACL.
* `block_public_policy` - (Optional) Whether Amazon S3 should block public bucket policies for this bucket. Defaults to `false`. Enabling this setting does not affect the existing bucket policy. When set to `true` causes Amazon S3 to:
  * Reject calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - (Optional) Whether Amazon S3 should ignore public ACLs for this bucket. Defaults to `false`. Enabling this setting does not affect the persistence of any existing ACLs and doesn't prevent new public ACLs from being set. When set to `true` causes Amazon S3 to:
  * Ignore all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - (Optional) Whether Amazon S3 should restrict public bucket policies for this bucket. Defaults to `false`. Enabling this setting does not affect the previously stored bucket policy, except that public and cross-account access within the public bucket policy, including non-public delegation to specific accounts, is blocked. When set to `true`:
  * Only the bucket owner and AWS Services can access this buckets if it has a public policy.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the S3 bucket the configuration is attached to

## Import

`aws_s3_bucket_public_access_block` can be imported by using the bucket name, e.g.

```
$ terraform import aws_s3_bucket_public_access_block.example my-bucket
```
