---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_user_group_membership"
description: |-
  Provides a resource for adding an IAM User to IAM Groups without conflicting
  with itself.
---

# Resource: aws_iam_user_group_membership

Provides a resource for adding an [IAM User][2] to [IAM Groups][1]. This
resource can be used multiple times with the same user for non-overlapping
groups.

To exclusively manage the users in a group, see the
[`aws_iam_group_membership` resource][3].

## Example usage

```hcl
resource "aws_iam_user_group_membership" "example1" {
  user = aws_iam_user.user1.name

  groups = [
    aws_iam_group.group1.name,
    aws_iam_group.group2.name,
  ]
}

resource "aws_iam_user_group_membership" "example2" {
  user = aws_iam_user.user1.name

  groups = [
    aws_iam_group.group3.name,
  ]
}

resource "aws_iam_user" "user1" {
  name = "user1"
}

resource "aws_iam_group" "group1" {
  name = "group1"
}

resource "aws_iam_group" "group2" {
  name = "group2"
}

resource "aws_iam_group" "group3" {
  name = "group3"
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The name of the [IAM User][2] to add to groups
* `groups` - (Required) A list of [IAM Groups][1] to add the user to

## Attributes Reference

* `user` - The name of the IAM User
* `groups` - The list of IAM Groups

[1]: /docs/providers/aws/r/iam_group.html
[2]: /docs/providers/aws/r/iam_user.html
[3]: /docs/providers/aws/r/iam_group_membership.html

## Import

IAM user group membership can be imported using the user name and group names separated by `/`.

```
$ terraform import aws_iam_user_group_membership.example1 user1/group1/group2
```
