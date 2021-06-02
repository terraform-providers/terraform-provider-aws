---
subcategory: "Amplify Console"
layout: "aws"
page_title: "AWS: aws_amplify_backend_environment"
description: |-
  Provides an Amplify Backend Environment resource.
---

# Resource: aws_amplify_backend_environment

Provides an Amplify Backend Environment resource.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name = "example"
}

resource "aws_amplify_backend_environment" "example" {
  app_id           = aws_amplify_app.example.id
  environment_name = "example"

  deployment_artifacts = "app-example-deployment"
  stack_name           = "amplify-app-example"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) The unique ID for an Amplify app.
* `environment_name` - (Required) The name for the backend environment.
* `deployment_artifacts` - (Optional) The name of deployment artifacts.
* `stack_name` - (Optional) The AWS CloudFormation stack name of a backend environment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for a backend environment that is part of an Amplify app.
* `id` - The unique ID of the Amplify backend environment.

## Import

Amplify backend environment can be imported using `app_id` and `environment_name`, e.g.

```
$ terraform import aws_amplify_backend_environment.example d2ypk4k47z8u6/example
```
