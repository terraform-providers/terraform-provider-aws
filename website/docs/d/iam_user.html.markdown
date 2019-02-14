---
layout: "aws"
page_title: "AWS: aws_iam_user"
sidebar_current: "docs-aws-datasource-iam-user"
description: |-
  Get information on a Amazon IAM user
---

# Data Source: aws_iam_user

This data source can be used to fetch information about a specific
IAM user. By using this data source, you can reference IAM user
properties without having to hard code ARNs or unique IDs as input.

## Example Usage

```hcl
data "aws_iam_user" "example" {
  user_name = "an_example_user_name"
}
```

## Argument Reference

* `user_name` - (Required) The friendly IAM user name to match.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) assigned by AWS for this user.
* `path` - Path in which this user was created.
* `permissions_boundary` - The ARN of the policy that is used to set the permissions boundary for the user.
* `user_id` - The unique ID assigned by AWS for this user.
* `user_name` - The name associated to this User
