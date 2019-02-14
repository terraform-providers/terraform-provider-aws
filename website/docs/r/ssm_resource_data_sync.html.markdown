---
layout: "aws"
page_title: "AWS: aws_ssm_resource_data_sync"
sidebar_current: "docs-aws-resource-ssm-resource-data-sync"
description: |-
  Provides a SSM resource data sync.
---

# aws_ssm_resource_data_sync

Provides a SSM resource data sync.

## Example Usage

```hcl
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test-bucket-1234"
  region = "us-east-1"
}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = "${aws_s3_bucket.hoge.bucket}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "SSMBucketPermissionsCheck",
            "Effect": "Allow",
            "Principal": {
                "Service": "ssm.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "arn:aws:s3:::tf-test-bucket-1234"
        },
        {
            "Sid": " SSMBucketDelivery",
            "Effect": "Allow",
            "Principal": {
                "Service": "ssm.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": ["arn:aws:s3:::tf-test-bucket-1234/*"],
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
    ]
}
EOF
}

resource "aws_ssm_resource_data_sync" "foo" {
  name = "foo"
  s3_destination = {
    bucket_name = "${aws_s3_bucket.hoge.bucket}"
    region      = "${aws_s3_bucket.hoge.region}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the configuration.
* `s3_destination` - (Required) Amazon S3 configuration details for the sync.

## s3_destination

`s3_destination` supports the following:

* `bucket_name` - (Required) Name of S3 bucket where the aggregated data is stored.
* `region` - (Required) Region with the bucket targeted by the Resource Data Sync.
* `kms_key_arn` - (Optional) ARN of an encryption key for a destination in Amazon S3.
* `prefix` - (Optional) Prefix for the bucket.
* `sync_format` - (Optional) A supported sync format. Only JsonSerDe is currently supported. Defaults to JsonSerDe.

## Import

SSM resource data sync can be imported using the `name`, e.g.

```sh
$ terraform import aws_ssm_resource_data_sync.example example-name
```
