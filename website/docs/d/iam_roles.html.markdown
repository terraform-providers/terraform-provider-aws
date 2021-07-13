---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_roles"
description: |-
  Get information about a set of IAM Roles.
---

# Data Source: aws_iam_roles

Use this data source to get ARNs and Names of IAM Roles that are created outside of the current Terraform state.

## Example Usage

### Retrieving all roles in an account

```terraform
data "aws_iam_roles" "roles" {}
```

### Retrieving roles by filter

Retrieving all IAM Roles whose role-name contains `project`

```terraform
data "aws_iam_roles" "roles" {
  filters = {
    name   = "role-name"
    values = ["*project*"]
  }
}
```

### Retrieving roles by path prefix

```terraform
data "aws_iam_roles" "roles" {
  path_prefix = "/custom-path"
}
```

### More examples

All IAM roles provisioned by AWS SSO in the account :

```terraform
data "aws_iam_roles" "roles" {
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
```

Specific IAM role provisioned by AWS SSO in the account :

```terraform
data "aws_iam_roles" "roles" {
  filters = {
    name   = "role-name"
    values = ["AWSReservedSSO_permission_set_name_*"]
  }
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
```

## Argument Reference

The following arguments are supported:

* `filters` - (Optional) One or more name/value pairs to use as filters. Filter names and values are case-sensitive. If using multiple filters for rules, the results include IAM Roles for which any combination of rules - not necessarily a single rule - match all filters.

  NOTICE: This filtering feature is not natively available in the [list-roles command of the AWS CLI][1]. Names of filters are based on the [Role structure][2] returned by the [list-roles command][1]. **Currently only the `role-name` filter is applicable.**

  `filters` are implemented using Glob matching as in EC2 APIs (See [Using Filtering](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html#Filtering_Resources_CLI) section).

* `path_prefix` - (Optional) The path prefix for filtering the results. For example, the prefix /application_abc/component_xyz/ gets all roles whose path starts with /application_abc/component_xyz/ . If it is not included, it defaults to a slash (/), listing all roles. For more details, check out [list-roles in the AWS CLI reference][1].

## Attributes Reference

* `arns` - ARNs of the matched IAM roles.
* `names` - Names of the matched IAM roles.

[1]: https://docs.aws.amazon.com/cli/latest/reference/iam/list-roles.html
[2]: https://docs.aws.amazon.com/cli/latest/reference/iam/list-roles.html#output
