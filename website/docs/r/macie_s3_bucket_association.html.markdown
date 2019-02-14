---
layout: "aws"
page_title: "AWS: aws_macie_s3_bucket_association"
sidebar_current: "docs-aws-macie-s3-bucket-association"
description: |-
  Associates an S3 resource with Amazon Macie for monitoring and data classification.
---

# aws_macie_s3_bucket_association

Associates an S3 resource with Amazon Macie for monitoring and data classification.

~> **NOTE:** Before using Amazon Macie for the first time it must be enabled manually. Instructions are [here](https://docs.aws.amazon.com/macie/latest/userguide/macie-setting-up.html#macie-setting-up-enable).

## Example Usage

```hcl
resource "aws_macie_s3_bucket_association" "example" {
  bucket_name = "tf-macie-example"
  prefix      = "data"

  classification_type {
    one_time = "FULL"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket_name` - (Required) The name of the S3 bucket that you want to associate with Amazon Macie.
* `classification_type` - (Optional) The configuration of how Amazon Macie classifies the S3 objects.
* `member_account_id` - (Optional) The ID of the Amazon Macie member account whose S3 resources you want to associate with Macie. If `member_account_id` isn't specified, the action associates specified S3 resources with Macie for the current master account.
* `prefix` - (Optional) Object key prefix identifying one or more S3 objects to which the association applies.

The `classification_type` object supports the following:

* `continuous` - (Optional) A string value indicating that Macie perform a one-time classification of all of the existing objects in the bucket.
The only valid value is the default value, `FULL`.
* `one_time` - (Optional) A string value indicating whether or not Macie performs a one-time classification of all of the existing objects in the bucket.
Valid values are `NONE` and `FULL`. Defaults to `NONE` indicating that Macie only classifies objects that are added after the association was created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association.
