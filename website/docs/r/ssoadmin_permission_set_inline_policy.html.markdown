---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_set_inline_policy"
description: |-
  Manages an IAM inline policy for a Single Sign-On (SSO) Permission Set
---

# Resource: aws_ssoadmin_permission_set_inline_policy

Provides an IAM inline policy for a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** AWS Single Sign-On (SSO) only supports one IAM inline policy per [`aws_ssoadmin_permission_set`](ssoadmin_permission_set.html) resource.
Creating or updating this resource will automatically [Provision the Permission Set](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_ProvisionPermissionSet.html) to apply the corresponding updates to all assigned accounts.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name         = "Example"
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

data "aws_iam_policy_document" "example" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "arn:aws:s3:::*",
    ]
  }
}

resource "aws_ssoadmin_permission_set_inline_policy" "example" {
  inline_policy      = data.aws_iam_policy_document.example.json
  instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `inline_policy` - (Required) The IAM inline policy to attach to a Permission Set.
* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `permission_set_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Permission Set.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Names (ARNs) of the Permission Set and SSO Instance, separated by a comma (`,`).

## Import

SSO Permission Set Inline Policies can be imported using the `permission_set_arn` and `instance_arn` separated by a comma (`,`) e.g.

```
$ terraform import aws_ssoadmin_permission_set_inline_policy.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
