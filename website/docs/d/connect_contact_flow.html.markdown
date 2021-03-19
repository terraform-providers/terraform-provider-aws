---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_contact_flow"
description: |-
  Provides details about a specific Amazon Connect Contact Flow.
---

# Data Source: aws_connect_contact_flow

Provides details about a specific Amazon Connect Contact Flow.

## Example Usage
By name

```hcl
data "aws_connect_contact_flow" "foo" {
  instance_id = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name        = "foo"
}
```

By contact_flow_id

```hcl
data "aws_connect_contact_flow" "foo" {
  instance_id     = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  contact_flow_id = "cccccccc-bbbb-cccc-dddd-111111111111"
}
```

## Argument Reference

~> **NOTE:** `instance_id` and one of either `name` or `contact_flow_id` is required.

The following arguments are supported:

* `instance_id` - (Required) Reference to the hosting Amazon Connect Instance

* `name` - (Optional) Returns information on a specific Contact Flow by name
* `contact_flow_id` - (Optional) Returns information on a specific Contact Flow by contact flow id

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Contact Flow.
* `description` - Specifies the description of the Contact Flow.
* `content` - Specifies the logic of the Contact Flow.
* `type` - Specifies the type of Contact Flow.
* `tags` - A the map of tags to assign to the Contact Flow.
