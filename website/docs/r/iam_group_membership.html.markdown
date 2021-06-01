---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_group_membership"
description: |-
  Provides a top level resource to manage IAM Group membership for IAM Users.
---

# Resource: aws_iam_group_membership

~> **WARNING:** Multiple aws_iam_group_membership resources with the same group name will produce inconsistent behavior!

Provides a top level resource to manage IAM Group membership for IAM Users. For
more information on managing IAM Groups or IAM Users, see [IAM Groups][1] or
[IAM Users][2]

~> **Note:** `aws_iam_group_membership` will conflict with itself if used more than once with the same group. To non-exclusively manage the users in a group, see the
[`aws_iam_user_group_membership` resource][3].

## Example Usage

```hcl
resource "aws_iam_group_membership" "team" {
  name = "tf-testing-group-membership"

  users = [
    aws_iam_user.user_one.name,
    aws_iam_user.user_two.name,
  ]

  group = aws_iam_group.group.name
}

resource "aws_iam_group" "group" {
  name = "test-group"
}

resource "aws_iam_user" "user_one" {
  name = "test-user"
}

resource "aws_iam_user" "user_two" {
  name = "test-user-two"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to identify the Group Membership
* `users` - (Required) A list of IAM User names to associate with the Group
* `group` – (Required) The IAM Group name to attach the list of `users` to

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name to identify the Group Membership
* `users` - list of IAM User names
* `group` – IAM Group name


[1]: /docs/providers/aws/r/iam_group.html
[2]: /docs/providers/aws/r/iam_user.html
[3]: /docs/providers/aws/r/iam_user_group_membership.html
