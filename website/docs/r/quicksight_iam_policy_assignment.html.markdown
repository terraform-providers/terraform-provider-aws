---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_iam_policy_assignment"
description: |-
  Manages a Resource QuickSight IAM Policy Assignment.
---

# Resource: aws_quicksight_iam_policy_assignment

Resource for managing QuickSight IAM Policy Assignment

## Example Usage

```hcl
resource "aws_quicksight_iam_policy_assignment" "example" {
  assignment_name = "exampleassignment"
  aws_account_id "123456789012"
  namespace = "default"
  assignment_status = "ENABLED"
  groups = ["group1", "group2", "group4"]
  users = ["username1", "username2"]
  policy_arn = "arn:aws:iam::123456789012:policy/example-policy"
}
```

## Argument Reference

The following arguments are supported:

* `assignment_name` - (Required) The name of the assignment. It must be unique within an AWS account. The name must contain only letters and numbers; special characters not allowed at the moment.
* `aws_account_id` - (Optional) The ID of the AWS account where you want to assign an IAM policy to QuickSight users or groups. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `namespace` - (Optional) The namespace that contains the assignment. Currently, you should set this to `default`.
* `assignment_status` - (Required) The status of the assignment. Possible values are as follows:
    - ENABLED - Anything specified in this assignment is used when creating
    - DISABLED - This assignment isn't used when creating the data source.
    - DRAFT - This assignment is an unfinished draft and isn't used when creating the data source.
* `groups` - (Optional) The QuickSight groups that you want to assign the policy to.
* `users` - (Optional) The QuickSight users that you want to assign the policy to.
* `policy_arn` - (Optional) The ARN for the IAM policy to apply to the QuickSight users and groups specified in this assignment.

## Timeouts

`aws_quicksight_iam_policy_assignment` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `45s`) How long to wait for the assignment to be created.

## Import

QuickSight IAM Policy Assignment can be imported using the aws account id, namespace and assignment name separated by `/`.

```
$ terraform import aws_quicksight_iam_policy_assignment.example 123456789012/default/exampleassignemnt
```
