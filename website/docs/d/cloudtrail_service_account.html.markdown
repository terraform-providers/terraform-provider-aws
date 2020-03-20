---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail_service_account"
description: |-
  Get AWS CloudTrail Service Account ID for storing trail data in S3.
---

# Data Source: aws_cloudtrail_service_account

Use this data source to get the Account ID of the [AWS CloudTrail Service Account](http://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html)
in a given region for the purpose of allowing CloudTrail to store trail data in S3.

## Example Usage

```hcl
data "aws_cloudtrail_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-cloudtrail-logging-test-bucket"
  force_destroy = true

  policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "Put bucket policy needed for trails",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_cloudtrail_service_account.main.arn}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::tf-cloudtrail-logging-test-bucket/*"
    },
    {
      "Sid": "Get bucket policy needed for trails",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_cloudtrail_service_account.main.arn}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:aws:s3:::tf-cloudtrail-logging-test-bucket"
    }
  ]
}
EOF
}
```

## Argument Reference

* `region` - (Optional) Name of the region whose AWS CloudTrail account ID is desired.
Defaults to the region from the AWS provider configuration.


## Attributes Reference

* `id` - The ID of the AWS CloudTrail service account in the selected region.
* `arn` - The ARN of the AWS CloudTrail service account in the selected region.
