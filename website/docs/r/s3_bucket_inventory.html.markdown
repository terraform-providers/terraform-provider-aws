---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_inventory"
description: |-
  Provides a S3 bucket inventory configuration resource.
---

# Resource: aws_s3_bucket_inventory

Provides a S3 bucket [inventory configuration](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-inventory.html) resource.

## Example Usage

### Add inventory configuration

```hcl
resource "aws_s3_bucket" "test" {
  bucket = "my-tf-test-bucket"
}

resource "aws_s3_bucket" "inventory" {
  bucket = "my-tf-inventory-bucket"
}

resource "aws_s3_bucket_inventory" "test" {
  bucket = aws_s3_bucket.test.id
  name   = "EntireBucketDaily"

  included_object_versions = "All"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.inventory.arn
    }
  }
}
```

### Add inventory configuration with S3 bucket object prefix

```hcl
resource "aws_s3_bucket" "test" {
  bucket = "my-tf-test-bucket"
}

resource "aws_s3_bucket" "inventory" {
  bucket = "my-tf-inventory-bucket"
}

resource "aws_s3_bucket_inventory" "test-prefix" {
  bucket = aws_s3_bucket.test.id
  name   = "DocumentsWeekly"

  included_object_versions = "All"

  schedule {
    frequency = "Daily"
  }

  filter {
    prefix = "documents/"
  }

  destination {
    bucket {
      format     = "ORC"
      bucket_arn = aws_s3_bucket.inventory.arn
      prefix     = "inventory"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the source bucket that inventory lists the objects for.
* `name` - (Required) Unique identifier of the inventory configuration for the bucket.
* `included_object_versions` - (Required) Object versions to include in the inventory list. Valid values: `All`, `Current`.
* `schedule` - (Required) Specifies the schedule for generating inventory results (documented below).
* `destination` - (Required) Contains information about where to publish the inventory results (documented below).
* `enabled` - (Optional, Default: `true`) Specifies whether the inventory is enabled or disabled.
* `filter` - (Optional) Specifies an inventory filter. The inventory only includes objects that meet the filter's criteria (documented below).
* `optional_fields` - (Optional) List of optional fields that are included in the inventory results.
Valid values: `Size`, `LastModifiedDate`, `StorageClass`, `ETag`, `IsMultipartUploaded`, `ReplicationStatus`, `EncryptionStatus`, `ObjectLockRetainUntilDate`, `ObjectLockMode`, `ObjectLockLegalHoldStatus`, `IntelligentTieringAccessTier`.

The `filter` configuration supports the following:

* `prefix` - (Optional) The prefix that an object must have to be included in the inventory results.

The `schedule` configuration supports the following:

* `frequency` - (Required) Specifies how frequently inventory results are produced. Valid values: `Daily`, `Weekly`.

The `destination` configuration supports the following:

* `bucket` - (Required) The S3 bucket configuration where inventory results are published (documented below).

The `bucket` configuration supports the following:

* `bucket_arn` - (Required) The Amazon S3 bucket ARN of the destination.
* `format` - (Required) Specifies the output format of the inventory results. Can be `CSV`, [`ORC`](https://orc.apache.org/) or [`Parquet`](https://parquet.apache.org/).
* `account_id` - (Optional) The ID of the account that owns the destination bucket. Recommended to be set to prevent problems if the destination bucket ownership changes.
* `prefix` - (Optional) The prefix that is prepended to all inventory results.
* `encryption` - (Optional) Contains the type of server-side encryption to use to encrypt the inventory (documented below).

The `encryption` configuration supports the following:

* `sse_kms` - (Optional) Specifies to use server-side encryption with AWS KMS-managed keys to encrypt the inventory file (documented below).
* `sse_s3` - (Optional) Specifies to use server-side encryption with Amazon S3-managed keys (SSE-S3) to encrypt the inventory file.

The `sse_kms` configuration supports the following:

* `key_id` - (Required) The ARN of the KMS customer master key (CMK) used to encrypt the inventory file.

## Import

S3 bucket inventory configurations can be imported using `bucket:inventory`, e.g.

```sh
$ terraform import aws_s3_bucket_inventory.my-bucket-entire-bucket my-bucket:EntireBucket
```
