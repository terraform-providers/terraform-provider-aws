---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_group_membership"
description: |-
  Manages a Resource QuickSight Group Membership.
---

# Resource: aws_quicksight_group_membership

Resource for managing QuickSight Group Membership

## Example Usage

```hcl
resource "aws_quicksight_group_membership" "example" {
	group_name  = "all-access-users"
	member_name = "john_smith"
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (Required) The name of the group in which the member will be added.
* `member_name` - (Required) The name of the member to add to the group.
* `aws_account_id` - (Optional) The ID for the AWS account that the group is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `namespace` - (Optional) The namespace. Currently, you should set this to `default`.

## Attributes Reference

All arguments above are exported.

## Import

QuickSight Group membership can be imported using the AWS account ID, namespace, group name and member name separated by `/`.

```
$ terraform import aws_quicksight_group.example 123456789123/default/all-access-users/john_smith
```
