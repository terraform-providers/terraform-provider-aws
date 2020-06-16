---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_user"
description: |-
  Get information on a Amazon QuickSight user
---

# Data Source: aws_quicksight_user

This data source can be used to fetch information about a specific
QuickSight user. By using this data source, you can reference QuickSight user
properties without having to hard code ARNs or unique IDs as input.

## Example Usage

```hcl
data "aws_quicksight_user" "example" {
  user_name 	 = "an_example_user_name"
  aws_account_id = "aws_account_id"
  namespace		 = "namespace"
}
```

## Argument Reference

* `user_name` - (Required) The name of the user that you want to match.
* `aws_account_id` - (Required) The ID for the AWS account that the user is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `namespace` - (Required) The namespace. Currently, you should set this to default.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) for the user.
* `active` - The active status of user. When you create an Amazon QuickSight user that’s not an IAM user or an Active Directory user, that user is inactive until they sign in and provide a password.
* `email` - The user's email address.
* `identity_type` - The type of identity authentication used by the user.
* `user_id` - The principal ID of the user.
* `user_role` - The Amazon QuickSight role for the user. The user role can be one of the following:.
    - READER: A user who has read-only access to dashboards.
    - AUTHOR: A user who can create data sources, datasets, analyses, and dashboards.
    - ADMIN: A user who is an author, who can also manage Amazon QuickSight settings.
