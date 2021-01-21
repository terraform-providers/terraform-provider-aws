---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_billing_service_account"
description: |-
  Get AWS Billing Service Account
---

# Data Source: aws_billing_service_account

Use this data source to get the Account ID of the [AWS Billing and Cost Management Service Account](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-getting-started.html#step-2) for the purpose of permitting in S3 bucket policy.

## Example Usage

```hcl
data "aws_billing_service_account" "main" {}

resource "aws_s3_bucket" "billing_logs" {
  bucket = "my-billing-tf-test-bucket"
  acl    = "private"

  policy = <<POLICY
{
  "Id": "Policy",
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetBucketAcl", "s3:GetBucketPolicy"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:s3:::my-billing-tf-test-bucket",
      "Principal": {
        "AWS": [
          "${data.aws_billing_service_account.main.arn}"
        ]
      }
    },
    {
      "Action": [
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:s3:::my-billing-tf-test-bucket/*",
      "Principal": {
        "AWS": [
          "${data.aws_billing_service_account.main.arn}"
        ]
      }
    }
  ]
}
POLICY
}
```


## Attributes Reference

* `id` - The ID of the AWS billing service account.
* `arn` - The ARN of the AWS billing service account.
