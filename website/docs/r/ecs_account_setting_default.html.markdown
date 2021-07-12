---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_account_setting_default"
description: |-
  Provides an ECS Default account setting.
---

# Resource: aws_ecs_account_setting_default

Provides an ECS default account setting for a specific ECS Resource name within a specific region. More information can be found on the [ECS Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-account-settings.html).

~> **NOTE:** The AWS API does not delete this resource, when you run a destroy this resource will be set to disabled.

## Example Usage

```hcl
resource "aws_ecs_account_setting_default" "test" {
  name  = "taskLongArnFormat"
  value = "enabled"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the account setting to set. Valid values are `serviceLongArnFormat`, `taskLongArnFormat`, `containerInstanceLongArnFormat`, `awsvpcTrunking` and `containerInsights`.
* `value` - (Required) The state of the setting current values are `enabled` or `disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the capacity provider.
* `prinicpal_arn` - The Amazon Resource Name (ARN) that identifies the account setting.

## Import

ECS Account Setting defaults can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_account_setting_default.example taskLongArnFormat
```
