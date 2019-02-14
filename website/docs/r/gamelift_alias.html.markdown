---
layout: "aws"
page_title: "AWS: aws_gamelift_alias"
sidebar_current: "docs-aws-resource-gamelift-alias"
description: |-
  Provides a Gamelift Alias resource.
---

# aws_gamelift_alias

Provides a Gamelift Alias resource.

## Example Usage

```hcl
resource "aws_gamelift_alias" "example" {
  name        = "example-alias"
  description = "Example Description"

  routing_strategy {
    message = "Example Message"
    type    = "TERMINAL"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the alias.
* `description` - (Optional) Description of the alias.
* `routing_strategy` - (Required) Specifies the fleet and/or routing type to use for the alias.

### Nested Fields

#### `routing_strategy`

* `fleet_id` - (Optional) ID of the Gamelift Fleet to point the alias to.
* `message` - (Optional) Message text to be used with the `TERMINAL` routing strategy.
* `type` - (Required) Type of routing strategy. e.g. `SIMPLE` or `TERMINAL`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Alias ID.
* `arn` - Alias ARN.

## Import

Gamelift Aliases can be imported using the ID, e.g.

```
$ terraform import aws_gamelift_alias.example <alias-id>
```
