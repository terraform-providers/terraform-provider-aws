---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_partition"
description: |-
  Get AWS partition identifier
---

# Data Source: aws_partition

Use this data source to lookup information about the current AWS partition in
which Terraform is working.

## Example Usage

```hcl
data "aws_partition" "current" {}

data "aws_iam_policy_document" "s3_policy" {
  statement {
    sid = "1"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::my-bucket",
    ]
  }
}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `dns_suffix` - Base DNS domain name for the current partition (e.g. `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
* `id` - Identifier of the current partition (e.g. `aws` in AWS Commercial, `aws-cn` in AWS China).
* `partition` - Identifier of the current partition (e.g. `aws` in AWS Commercial, `aws-cn` in AWS China).
* `reverse_dns_prefix` - Prefix of service names (e.g. `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).
