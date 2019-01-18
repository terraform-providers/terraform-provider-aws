---
layout: "aws"
page_title: "AWS: aws_iam_group_policy_attachment"
sidebar_current: "docs-aws-resource-iam-group-policy-attachment"
description: |-
  Attaches a Managed IAM Policy to an IAM group
---

# aws_iam_group_policy_attachment

Attaches a Managed IAM Policy to an IAM group

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_policy_attachment` resource and will permanently show a difference if both are defined.

## Example Usage

```hcl
resource "aws_iam_group" "group" {
  name = "test-group"
}

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"
  policy      = "" # insert policy here
}

resource "aws_iam_group_policy_attachment" "test-attach" {
  group      = "${aws_iam_group.group.name}"
  policy_arn = "${aws_iam_policy.policy.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `group`  (Required) - The group the policy should be applied to
* `policy_arn`  (Required) - The ARN of the policy you want to apply

## Import

IAM group policy attachments can be imported using the group name and policy arn separated by `/`.

```
$ terraform import aws_iam_group_policy_attachment.test-attach test-group/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy
```
