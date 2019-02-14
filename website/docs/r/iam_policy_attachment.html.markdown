---
layout: "aws"
page_title: "AWS: aws_iam_policy_attachment"
sidebar_current: "docs-aws-resource-iam-policy-attachment"
description: |-
  Attaches a Managed IAM Policy to user(s), role(s), and/or group(s)
---

# aws_iam_policy_attachment

Attaches a Managed IAM Policy to user(s), role(s), and/or group(s)

!> **WARNING:** The aws_iam_policy_attachment resource creates **exclusive** attachments of IAM policies. Across the entire AWS account, all of the users/roles/groups to which a single policy is attached must be declared by a single aws_iam_policy_attachment resource. This means that even any users/roles/groups that have the attached policy via any other mechanism (including other Terraform resources) will have that attached policy revoked by this resource. Consider `aws_iam_role_policy_attachment`, `aws_iam_user_policy_attachment`, or `aws_iam_group_policy_attachment` instead. These resources do not enforce exclusive attachment of an IAM policy.

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_group_policy_attachment`, `aws_iam_role_policy_attachment`, and `aws_iam_user_policy_attachment` resources and will permanently show a difference if both are defined.

## Example Usage

```hcl
resource "aws_iam_user" "user" {
  name = "test-user"
}

resource "aws_iam_role" "role" {
  name = "test-role"
}

resource "aws_iam_group" "group" {
  name = "test-group"
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = "" # insert policy here
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment"
  users      = ["${aws_iam_user.user.name}"]
  roles      = ["${aws_iam_role.role.name}"]
  groups     = ["${aws_iam_group.group.name}"]
  policy_arn = "${aws_iam_policy.policy.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `name`    (Required) - The name of the attachment. This cannot be an empty string.
* `users`   (Optional) - The user(s) the policy should be applied to
* `roles`   (Optional) - The role(s) the policy should be applied to
* `groups`  (Optional) - The group(s) the policy should be applied to
* `policy_arn`  (Required) - The ARN of the policy you want to apply

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The policy's ID.
* `name` - The name of the attachment.
