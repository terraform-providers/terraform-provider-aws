---
subcategory: "Step Function (SFN)"
layout: "aws"
page_title: "AWS: aws_sfn_state_machine"
description: |-
  Get information on an Amazon Step Function State Machine
---

# Data Source: aws_sfn_state_machine

Use this data source to get the ARN of a State Machine in AWS Step
Function (SFN). By using this data source, you can reference a
state machine without having to hard code the ARNs as input.

## Example Usage

```hcl
data "aws_sfn_state_machine" "example" {
  name = "an_example_sfn_name"
}
```

## Argument Reference

* `name` - (Required) The friendly name of the state machine to match.

## Attributes Reference

* `id` - Set to the ARN of the found state machine, suitable for referencing in other resources that support State Machines.
* `arn` - Set to the arn of the state function.
* `role_arn` - Set to the role_arn used by the state function.
* `definition` - Set to the state machine definition.
* `creation_date` - The date the state machine was created.
* `status` - Set to the current status of the state machine.
