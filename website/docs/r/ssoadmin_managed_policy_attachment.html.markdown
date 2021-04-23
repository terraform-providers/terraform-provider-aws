---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_managed_policy_attachment"
description: |-
  Manages an IAM managed policy for a Single Sign-On (SSO) Permission Set
---

# Resource: aws_ssoadmin_managed_policy_attachment

Provides an IAM managed policy for a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** Creating this resource will automatically [Provision the Permission Set](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_ProvisionPermissionSet.html) to apply the corresponding updates to all assigned accounts.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name         = "Example"
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "example" {
  instance_arn       = aws_ssoadmin_permission_set.example.instance_arn
  managed_policy_arn = "arn:aws:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `managed_policy_arn` - (Required, Forces new resource) The IAM managed policy Amazon Resource Name (ARN) to be attached to the Permission Set.
* `permission_set_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the Permission Set.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Names (ARNs) of the Managed Policy, Permission Set, and SSO Instance, separated by a comma (`,`).
* `managed_policy_name` - The name of the IAM Managed Policy.

## Import

SSO Managed Policy Attachments can be imported using the `managed_policy_arn`, `permission_set_arn`, and `instance_arn` separated by a comma (`,`) e.g.

```
$ terraform import aws_ssoadmin_managed_policy_attachment.example arn:aws:iam::aws:policy/AlexaForBusinessDeviceSetup,arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
