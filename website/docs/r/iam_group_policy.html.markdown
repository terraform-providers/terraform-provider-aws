---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_group_policy"
description: |-
  Provides an IAM policy attached to a group.
---

# Resource: aws_iam_group_policy

Provides an IAM policy attached to a group.

## Example Usage

```hcl
resource "aws_iam_group_policy" "my_developer_policy" {
  name  = "my_developer_policy"
  group = aws_iam_group.my_developers.name

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

resource "aws_iam_group" "my_developers" {
  name = "developers"
  path = "/users/"
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy)
* `name` - (Optional) The name of the policy. If omitted, Terraform will
assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified
  prefix. Conflicts with `name`.
* `group` - (Required) The IAM group to attach to the policy.

## Attributes Reference

* `id` - The group policy ID.
* `group` - The group to which this policy applies.
* `name` - The name of the policy.
* `policy` - The policy document attached to the group.

## Import

IAM Group Policies can be imported using the `group_name:group_policy_name`, e.g.

```
$ terraform import aws_iam_group_policy.mypolicy group_of_mypolicy_name:mypolicy_name
```
