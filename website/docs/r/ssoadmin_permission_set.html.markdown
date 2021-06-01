---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_set"
description: |-
  Manages a Single Sign-On (SSO) Permission Set
---

# Resource: aws_ssoadmin_permission_set

Provides a Single Sign-On (SSO) Permission Set resource

~> **NOTE:** Updating this resource will automatically [Provision the Permission Set](https://docs.aws.amazon.com/singlesignon/latest/APIReference/API_ProvisionPermissionSet.html) to apply the corresponding updates to all assigned accounts.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_permission_set" "example" {
  name             = "Example"
  description      = "An example"
  instance_arn     = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  relay_state      = "https://s3.console.aws.amazon.com/s3/home?region=us-east-1#"
  session_duration = "PT2H"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) The description of the Permission Set.
* `instance_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the SSO Instance under which the operation will be executed.
* `name` - (Required, Forces new resource) The name of the Permission Set.
* `relay_state` - (Optional) The relay state URL used to redirect users within the application during the federation authentication process.
* `session_duration` - (Optional) The length of time that the application user sessions are valid in the ISO-8601 standard. Default: `PT1H`.
* `tags` - (Optional) Key-value map of resource tags.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Permission Set.
* `id` - The Amazon Resource Names (ARNs) of the Permission Set and SSO Instance, separated by a comma (`,`).
* `created_date` - The date the Permission Set was created in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).

## Import

SSO Permission Sets can be imported using the `arn` and `instance_arn` separated by a comma (`,`) e.g.

```
$ terraform import aws_ssoadmin_permission_set.example arn:aws:sso:::permissionSet/ssoins-2938j0x8920sbj72/ps-80383020jr9302rk,arn:aws:sso:::instance/ssoins-2938j0x8920sbj72
```
