---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_role"
description: |-
  Provides an IAM role.
---

# Resource: aws_iam_role

Provides an IAM role.

~> **NOTE:** If policies are attached to the role via the [`aws_iam_policy_attachment` resource](/docs/providers/aws/r/iam_policy_attachment.html) and you are modifying the role `name` or `path`, the `force_detach_policies` argument must be set to `true` and applied before attempting the operation otherwise you will encounter a `DeleteConflict` error. The [`aws_iam_role_policy_attachment` resource (recommended)](/docs/providers/aws/r/iam_role_policy_attachment.html) does not have this requirement.

~> **NOTE:** If you use this resource's `managed_policy_arns` argument or `inline_policy` configuration blocks, this resource will take over exclusive management of the role's respective policy types (e.g., both policy types if both arguments are used). These arguments are incompatible with other ways of managing a role's policies, such as [`aws_iam_policy_attachment`](/docs/providers/aws/r/iam_policy_attachment.html), [`aws_iam_role_policy_attachment`](/docs/providers/aws/r/iam_role_policy_attachment.html), and [`aws_iam_role_policy`](/docs/providers/aws/r/iam_role_policy.html). If you attempt to manage a role's policies by multiple means, you will get resource cycling and/or errors.

## Example Usage

### Basic Example

```terraform
resource "aws_iam_role" "test_role" {
  name = "test_role"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    tag-key = "tag-value"
  }
}
```

### Example of Using Data Source for Assume Role Policy

```terraform
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
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}
```

### Example of Exclusive Inline Policies

This example creates an IAM role with two inline IAM policies. If someone adds another inline policy out-of-band, on the next apply, Terraform will remove that policy. If someone deletes these policies out-of-band, Terraform will recreate them.

```terraform
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role_policy.json # (not shown)

  inline_policy {
    name = "my_inline_policy"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action   = ["ec2:Describe*"]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }

  inline_policy {
    name   = "policy-8675309"
    policy = data.aws_iam_policy_document.inline_policy.json
  }
}

data "aws_iam_policy_document" "inline_policy" {
  statement {
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}
```

### Example of Removing Inline Policies

This example creates an IAM role with what appears to be empty IAM `inline_policy` argument instead of using `inline_policy` as a configuration block. The result is that if someone were to add an inline policy out-of-band, on the next apply, Terraform will remove that policy.


```terraform
resource "aws_iam_role" "example" {
  name               = "yak_role"
  assume_role_policy = data.aws_iam_policy_document.instance_assume_role_policy.json # (not shown)

  inline_policy {}
}
```

### Example of Exclusive Managed Policies

This example creates an IAM role and attaches two managed IAM policies. If someone attaches another managed policy out-of-band, on the next apply, Terraform will detach that policy. If someone detaches these policies out-of-band, Terraform will attach them again.

```terraform
resource "aws_iam_role" "example" {
  name                = "yak_role"
  assume_role_policy  = data.aws_iam_policy_document.instance_assume_role_policy.json # (not shown)
  managed_policy_arns = [aws_iam_policy.policy_one.arn, aws_iam_policy.policy_two.arn]
}

resource "aws_iam_policy" "policy_one" {
  name = "policy-618033"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["ec2:Describe*"]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_policy" "policy_two" {
  name = "policy-381966"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["s3:ListAllMyBuckets", "s3:ListBucket", "s3:HeadBucket"]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
```

### Example of Removing Managed Policies

This example creates an IAM role with an empty `managed_policy_arns` argument. If someone attaches a policy out-of-band, on the next apply, Terraform will detach that policy.

```terraform
resource "aws_iam_role" "example" {
  name                = "yak_role"
  assume_role_policy  = data.aws_iam_policy_document.instance_assume_role_policy.json # (not shown)
  managed_policy_arns = []
}
```

## Argument Reference

The following argument is required:

* `assume_role_policy` - (Required) Policy that grants an entity permission to assume the role.

~> **NOTE:** The `assume_role_policy` is very similar to but slightly different than a standard IAM policy and cannot use an `aws_iam_policy` resource.  However, it _can_ use an `aws_iam_policy_document` [data source](/docs/providers/aws/d/iam_policy_document.html). See the example above of how this works.

The following arguments are optional:

* `description` - (Optional) Description of the role.
* `force_detach_policies` - (Optional) Whether to force detaching any policies the role has before destroying it. Defaults to `false`.
* `inline_policy` - (Optional) Configuration block defining an exclusive set of IAM inline policies associated with the IAM role. Defined below. If no blocks are configured, Terraform will ignore any managing any inline policies in this resource. Configuring one empty block (i.e., `inline_policy {}`) will cause Terraform to remove _all_ inline policies.
* `managed_policy_arns` - (Optional) Set of exclusive IAM managed policy ARNs to attach to the IAM role. If this attribute is not configured, Terraform will ignore policy attachments to this resource. When configured, Terraform will align the role's managed policy attachments with this set by attaching or detaching managed policies. Configuring an empty set (i.e., `managed_policy_arns = []`) will cause Terraform to remove _all_ managed policy attachments.
* `max_session_duration` - (Optional) Maximum session duration (in seconds) that you want to set for the specified role. If you do not specify a value for this setting, the default maximum of one hour is applied. This setting can have a value from 1 hour to 12 hours.
* `name` - (Optional, Forces new resource) Friendly name of the role. If omitted, Terraform will assign a random, unique name. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) for more information.
* `name_prefix` - (Optional, Forces new resource) Creates a unique friendly name beginning with the specified prefix. Conflicts with `name`.
* `path` - (Optional) Path to the role. See [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) for more information.
* `permissions_boundary` - (Optional) ARN of the policy that is used to set the permissions boundary for the role.
* `tags` - Key-value mapping of tags for the IAM role. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### inline_policy

This configuration block supports the following:

* `name` - (Required) Name of the role policy.
* `policy` - (Required) Policy document as a JSON formatted string. For more information about building IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://www.terraform.io/docs/providers/aws/guides/iam-policy-documents.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) specifying the role.
* `create_date` - Creation date of the IAM role.
* `id` - Name of the role.
* `name` - Name of the role.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `unique_id` - Stable and unique string identifying the role.

## Import

IAM Roles can be imported using the `name`, e.g.

```
$ terraform import aws_iam_role.developer developer_name
```
