---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_user"
description: |-
  Manages a Resource QuickSight User.
---

# Resource: aws_quicksight_user

Resource for managing QuickSight User

## Example Usage

```hcl
resource "aws_quicksight_user" "example" {
  user_name     = "an-author"
  email         = "author@example.com"
  identity_type = "IAM"
  user_role     = "AUTHOR"
}
```

## Argument Reference

The following arguments are supported:


* `email` - (Required) The email address of the user that you want to register.
* `identity_type` - (Required) Amazon QuickSight supports several ways of managing the identity of users. This parameter accepts either  `IAM` or `QUICKSIGHT`.
* `user_role` - (Required) The Amazon QuickSight role of the user. The user role can be one of the following: `READER`, `AUTHOR`, or `ADMIN`
* `user_name` - (Optional) The Amazon QuickSight user name that you want to create for the user you are registering.
* `aws_account_id` - (Optional) The ID for the AWS account that the user is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `iam_arn` - (Optional) The ARN of the IAM user or role that you are registering with Amazon QuickSight.
* `namespace`  - (Optional) The namespace. Currently, you should set this to `default`.
* `session_name` - (Optional) The name of the IAM session to use when assuming roles that can embed QuickSight dashboards.

## Attributes Reference

All above attributes except for `session_name` and `identity_type` are exported as well as:

* `arn` - Amazon Resource Name (ARN) of the user

## Import

Importing is currently not supported on this resource.
