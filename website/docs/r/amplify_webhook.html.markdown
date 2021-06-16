---
subcategory: "Amplify Console"
layout: "aws"
page_title: "AWS: aws_amplify_webhook"
description: |-
  Provides an Amplify Webhook resource.
---

# Resource: aws_amplify_webhook

Provides an Amplify Webhook resource.

## Example Usage

```terraform
resource "aws_amplify_app" "example" {
  name = "app"
}

resource "aws_amplify_branch" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = "master"
}

resource "aws_amplify_webhook" "master" {
  app_id      = aws_amplify_app.example.id
  branch_name = aws_amplify_branch.master.branch_name
  description = "triggermaster"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) The unique ID for an Amplify app.
* `branch_name` - (Required) The name for a branch that is part of the Amplify app.
* `description` - (Optional) The description for a webhook.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) for the webhook.
* `url` - The URL of the webhook.

## Import

Amplify webhook can be imported using a webhook ID, e.g.

```
$ terraform import aws_amplify_webhook.master a26b22a0-748b-4b57-b9a0-ae7e601fe4b1
```
