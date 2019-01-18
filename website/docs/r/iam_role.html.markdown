---
layout: "aws"
page_title: "AWS: aws_iam_role"
sidebar_current: "docs-aws-resource-iam-role"
description: |-
  Provides an IAM role.
---

# aws_iam_role

Provides an IAM role.

## Example Usage

```hcl
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

  tags = {
      tag-key = "tag-value"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the role. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `assume_role_policy` - (Required) The policy that grants an entity permission to assume the role.

~> **NOTE:** This `assume_role_policy` is very similar but slightly different than just a standard IAM policy and cannot use an `aws_iam_policy` resource.  It _can_ however, use an `aws_iam_policy_document` [data source](https://www.terraform.io/docs/providers/aws/d/iam_policy_document.html), see example below for how this could work.

* `force_detach_policies` - (Optional) Specifies to force detaching any policies the role has before destroying it. Defaults to `false`.
* `path` - (Optional) The path to the role.
  See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) for more information.
* `description` - (Optional) The description of the role.

* `max_session_duration` - (Optional) The maximum session duration (in seconds) that you want to set for the specified role. If you do not specify a value for this setting, the default maximum of one hour is applied. This setting can have a value from 1 hour to 12 hours.
* `permissions_boundary` - (Optional) The ARN of the policy that is used to set the permissions boundary for the role.
* `tags` - Key-value mapping of tags for the IAM role

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) specifying the role.
* `create_date` - The creation date of the IAM role.
* `description` - The description of the role.
* `id` - The name of the role.
* `name` - The name of the role.
* `unique_id` - The stable and unique string identifying the role.

## Example of Using Data Source for Assume Role Policy

```hcl
data "aws_iam_policy_document" "instance-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "instance" {
  name               = "instance_role"
  path               = "/system/"
  assume_role_policy = "${data.aws_iam_policy_document.instance-assume-role-policy.json}"
}
```

## Import

IAM Roles can be imported using the `name`, e.g.

```
$ terraform import aws_iam_role.developer developer_name
```
