---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_group"
description: |-
  Get information on a Amazon IAM group
---

# Data Source: aws_iam_group

This data source can be used to fetch information about a specific
IAM group. By using this data source, you can reference IAM group
properties without having to hard code ARNs as input.

## Example Usage

```hcl
data "aws_iam_group" "example" {
  group_name = "an_example_group_name"
}
```

## Argument Reference

* `group_name` - (Required) The friendly IAM group name to match.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) specifying the group.

* `path` - The path to the group.

* `group_id` - The stable and unique string identifying the group.

* `users` - List of objects containing group member information. See supported fields below.

### `users`

* `arn` - The Amazon Resource Name (ARN) specifying the iam user.

* `user_id` - The stable and unique string identifying the iam user.

* `user_name` - The name of the iam user.

* `path` - The path to the iam user.
