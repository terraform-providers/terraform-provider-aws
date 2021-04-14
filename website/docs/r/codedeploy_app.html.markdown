---
subcategory: "CodeDeploy"
layout: "aws"
page_title: "AWS: aws_codedeploy_app"
description: |-
  Provides a CodeDeploy application.
---

# Resource: aws_codedeploy_app

Provides a CodeDeploy application to be used as a basis for deployments

## Example Usage

### ECS Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "ECS"
  name             = "example"
}
```

### Lambda Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "Lambda"
  name             = "example"
}
```

### Server Application

```terraform
resource "aws_codedeploy_app" "example" {
  compute_platform = "Server"
  name             = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the application.
* `compute_platform` - (Optional) The compute platform can either be `ECS`, `Lambda`, or `Server`. Default is `Server`.
* `tags` - (Optional) Key-value map of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the CodeDeploy application.
* `application_id` - The application ID.
* `id` - Amazon's assigned ID for the application.
* `name` - The application's name.
* `github_account_name` - The name for a connection to a GitHub account.
* `linked_to_github` - Whether the user has authenticated with GitHub for the specified application.

## Import

CodeDeploy Applications can be imported using the `name`, e.g.

```
$ terraform import aws_codedeploy_app.example my-application
```
