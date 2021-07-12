---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_assume_role_policy"
description: |-
  Configures the assume role policy for an IAM role.
---

# Resource: aws_iam_assume_role_policy

Configures the assume role policy for an IAM role.  This is useful in cases where a terraform user has restricted access to AWS accounts that prevent them from creating IAM roles.

## Example Usage

```hcl
resource "aws_iam_assume_role_policy" "test_policy" {
  role = "${aws_iam_role.test_role.id}"

  policy = <<EOF
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

* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](/docs/providers/aws/guides/iam-policy-documents.html)
* `role` - (Required) The IAM role to attach to the policy.

## Attributes Reference

* `id` - The name of the role associated with the policy.
* `policy` - The policy document attached to the role.
* `role` - The name of the role associated with the policy.