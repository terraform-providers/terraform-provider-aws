---
layout: "aws"
page_title: "AWS: aws_organizations_policy"
sidebar_current: "docs-aws-resource-organizations-policy"
description: |-
  Provides a resource to manage an AWS Organizations policy.
---

# aws_organizations_policy

Provides a resource to manage an [AWS Organizations policy](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies.html).

## Example Usage

```hcl
resource "aws_organizations_policy" "example" {
  name = "example"

  content = <<CONTENT
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
CONTENT
}
```

## Argument Reference

The following arguments are supported:

* `content` - (Required) The policy content to add to the new policy. For example, if you create a [service control policy (SCP)](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scp.html), this string must be JSON text that specifies the permissions that admins in attached accounts can delegate to their users, groups, and roles. For more information about the SCP syntax, see the [Service Control Policy Syntax documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_reference_scp-syntax.html).
* `name` - (Required) The friendly name to assign to the policy.
* `description` - (Optional) A description to assign to the policy.
* `type` - (Optional) The type of policy to create. Currently, the only valid value is `SERVICE_CONTROL_POLICY` (SCP).

## Attribute Reference

* `id` - The unique identifier (ID) of the policy.
* `arn` - Amazon Resource Name (ARN) of the policy.

## Import

`aws_organizations_policy` can be imported by using the policy ID, e.g.

```
$ terraform import aws_organizations_policy.example p-12345678
```
