---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_policy"
description: |-
  Get information on a Amazon IAM policy
---

# Data Source: aws_iam_policy

This data source can be used to fetch information about a specific
IAM policy.

## Example Usage

### By ARN

```terraform
data "aws_iam_policy" "example" {
  arn = "arn:aws:iam::123456789012:policy/UsersManageOwnCredentials"
}
```

### By Name

```terraform
data "aws_iam_policy" "example" {
  name = "test_policy"
}
```

## Argument Reference

* `arn` - (Optional) The ARN of the IAM policy.
* `name` - (Optional) The name of the IAM policy.
* `path_prefix` - (Optional) The prefix of the path to the IAM policy. Defaults to a slash (`/`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `path` - The path to the policy.
* `description` - The description of the policy.
* `policy` - The policy document of the policy.
* `policy_id` - The policy's ID.
* `tags` - Key-value mapping of tags for the IAM Policy.
