---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Get information on a Amazon IAM policy
---

# aws_iam_policy

This data source can be used to fetch information about a specific
IAM policy.

## Example Usage

```hcl
data "aws_iam_policy" "example" {
  arn = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
}
```

## Argument Reference

* `arn` - (Required) ARN of the IAM policy.

## Attributes Reference

* `name` - The name of the IAM policy.
* `arn` - The Amazon Resource Name (ARN) specifying the policy.
* `path` - The path to the policy.
* `description` - The description of the policy.
* `policy` - The policy document of the policy.

