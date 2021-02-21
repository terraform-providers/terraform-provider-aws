---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_permission_set"
description: |-
  Get information on a Single Sign-On (SSO) Permission Set.
---

# Data Source: aws_ssoadmin_permission_set

Use this data source to get a Single Sign-On (SSO) Permission Set.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

data "aws_ssoadmin_permission_set" "example" {
  instance_arn = tolist(data.aws_ssoadmin_instances.example.arns)[0]
  name         = "Example"
}

output "arn" {
  value = data.aws_ssoadmin_permission_set.example.arn
}
```

## Argument Reference

The following arguments are supported:

~> **NOTE:** Either `arn` or `name` must be configured.

* `arn` - (Optional) The Amazon Resource Name (ARN) of the permission set.
* `instance_arn` - (Required) The Amazon Resource Name (ARN) of the SSO Instance associated with the permission set.
* `name` - (Optional) The name of the SSO Permission Set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the Permission Set.
* `description` - The description of the Permission Set.
* `relay_state` - The relay state URL used to redirect users within the application during the federation authentication process.
* `session_duration` - The length of time that the application user sessions are valid in the ISO-8601 standard.
* `tags` - Key-value map of resource tags.
