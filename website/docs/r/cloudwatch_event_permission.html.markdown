---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_permission"
description: |-
  Provides a resource to create an EventBridge permission to support cross-account events in the current account default event bus.
---

# Resource: aws_cloudwatch_event_permission

Provides a resource to create an EventBridge permission to support cross-account events in the current account default event bus.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

### Account Access

```hcl
resource "aws_cloudwatch_event_permission" "DevAccountAccess" {
  principal    = "123456789012"
  statement_id = "DevAccountAccess"
}
```

### Organization Access

```hcl
resource "aws_cloudwatch_event_permission" "OrganizationAccess" {
  principal    = "*"
  statement_id = "OrganizationAccess"

  condition {
    key   = "aws:PrincipalOrgID"
    type  = "StringEquals"
    value = aws_organizations_organization.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `principal` - (Required) The 12-digit AWS account ID that you are permitting to put events to your default event bus. Specify `*` to permit any account to put events to your default event bus, optionally limited by `condition`.
* `statement_id` - (Required) An identifier string for the external account that you are granting permissions to.
* `action` - (Optional) The action that you are enabling the other account to perform. Defaults to `events:PutEvents`.
* `condition` - (Optional) Configuration block to limit the event bus permissions you are granting to only accounts that fulfill the condition. Specified below.
* `event_bus_name` - (Optional) The event bus to set the permissions on. If you omit this, the permissions are set on the `default` event bus.

### condition

* `key` - (Required) Key for the condition. Valid values: `aws:PrincipalOrgID`.
* `type` - (Required) Type of condition. Value values: `StringEquals`.
* `value` - (Required) Value for the key.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The statement ID of the EventBridge permission.

## Import

EventBridge permissions can be imported using the `event_bus_name/statement_id` (if you omit `event_bus_name`, the `default` event bus will be used), e.g.

```shell
$ terraform import aws_cloudwatch_event_permission.DevAccountAccess example-event-bus/DevAccountAccess
```
