---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_policy"
description: |-
    Provides IAM policy of an S3 bucket
---

# Data Source: aws_s3_bucket_policy
The bucket-policy data source returns IAM policy of an S3 bucket.

## Example Usage

The following example retrieves IAM policy of a specified S3 bucket.

```hcl
data "aws_s3_bucket_policy" "policy" {
  bucket = "example-bucket-name"
}

output "foo" {
  value = data.aws_s3_bucket_policy.policy
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to read the policy from.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy` - IAM policy attached to the S3 bucket
