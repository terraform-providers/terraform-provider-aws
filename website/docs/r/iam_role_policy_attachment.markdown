---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_role_policy_attachment"
description: |-
  Attaches a Managed IAM Policy to an IAM role
---

# Resource: aws_iam_role_policy_attachment

Attaches a Managed IAM Policy to an IAM role

~> **NOTE:** The usage of this resource conflicts with the `aws_iam_policy_attachment` resource and will permanently show a difference if both are defined.

~> **NOTE:** For a given role, this resource is incompatible with using the [`aws_iam_role` resource](/docs/providers/aws/r/iam_role.html) `managed_policy_arns` argument. When using that argument and this resource, both will attempt to manage the role's managed policy attachments and Terraform will show a permanent difference.

## Example Usage

```terraform
resource "aws_iam_role" "role" {
  name = "test-role"

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

resource "aws_iam_policy" "policy" {
  name        = "test-policy"
  description = "A test policy"

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

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy.arn
}
```

## Argument Reference

The following arguments are supported:

* `role`  (Required) - The name of the IAM role to which the policy should be applied
* `policy_arn` (Required) - The ARN of the policy you want to apply

## Attributes Reference

No additional attributes are exported.

## Import

IAM role policy attachments can be imported using the role name and policy arn separated by `/`.

```
$ terraform import aws_iam_role_policy_attachment.test-attach test-role/arn:aws:iam::xxxxxxxxxxxx:policy/test-policy
```
