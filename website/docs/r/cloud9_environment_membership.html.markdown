---
subcategory: "Cloud9"
layout: "aws"
page_title: "AWS: aws_cloud9_environment_membership"
description: |-
  Provides an environment member to an AWS Cloud9 development environment.
---

# Resource: aws_cloud9_environment_membership

Provides an environment member to an AWS Cloud9 development environment.

## Example Usage

```terraform
resource "aws_cloud9_environment_ec2" "test" {
  instance_type = "t2.micro"
  name          = "some-env"
}

resource "aws_iam_user" "test" {
  name = "some-user"
}

resource "aws_cloud9_environment_membership" "test" {
  environment_id = aws_cloud9_environment_ec2.test.id
  permissions    = "read-only"
  user_arn       = aws_iam_user.test.arn
}
```

## Argument Reference

The following arguments are supported:

* `environment_id` - (Required) The ID of the environment that contains the environment member you want to add.
* `permissions` - (Required) The type of environment member permissions you want to associate with this environment member. Allowed values are `read-only` and `read-write` .
* `user_arn` - (Required) The Amazon Resource Name (ARN) of the environment member you want to add.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the environment membership.
* `user_id` - he user ID in AWS Identity and Access Management (AWS IAM) of the environment member.

## Import

Cloud9 environment membership can be imported using the `environment-id#user-arn`, e.g.

```
$ terraform import aws_cloud9_environment_membership.test environment-id#user-arn
```