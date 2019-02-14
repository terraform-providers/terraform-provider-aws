---
layout: "aws"
page_title: "AWS: aws_iam_role_policy"
sidebar_current: "docs-aws-resource-iam-role-policy"
description: |-
  Provides an IAM role policy.
---

# aws_iam_role_policy

Provides an IAM role policy.

## Example Usage

```hcl
resource "aws_iam_role_policy" "test_policy" {
  name = "test_policy"
  role = "${aws_iam_role.test_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test_role" {
  name = "test_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the role policy. If omitted, Terraform will
assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified
  prefix. Conflicts with `name`.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](/docs/providers/aws/guides/iam-policy-documents.html)
* `role` - (Required) The IAM role to attach to the policy.

## Attributes Reference

* `id` - The role policy ID, in the form of `role_name:role_policy_name`.
* `name` - The name of the policy.
* `policy` - The policy document attached to the role.
* `role` - The name of the role associated with the policy.

## Import

IAM Role Policies can be imported using the `role_name:role_policy_name`, e.g.

```
$ terraform import aws_iam_role_policy.mypolicy role_of_mypolicy_name:mypolicy_name
```
