---
layout: "aws"
page_title: "AWS: aws_iam_role"
sidebar_current: "docs-aws-datasource-iam-role"
description: |-
  Get information on a Amazon IAM role
---

# Data Source: aws_iam_role

This data source can be used to fetch information about a specific
IAM role. By using this data source, you can reference IAM role
properties without having to hard code ARNs as input.

## Example Usage

```hcl
data "aws_iam_role" "example" {
  name = "an_example_role_name"
}
```

## Argument Reference

* `name` - (Required) The friendly IAM role name to match.

## Attributes Reference

* `id` - The friendly IAM role name to match.
* `arn` - The Amazon Resource Name (ARN) specifying the role.
* `assume_role_policy` - The policy document associated with the role.
* `path` - The path to the role.
* `permissions_boundary` - The ARN of the policy that is used to set the permissions boundary for the role.
* `unique_id` - The stable and unique string identifying the role.
