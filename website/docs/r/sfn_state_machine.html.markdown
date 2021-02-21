---
subcategory: "Step Function (SFN)"
layout: "aws"
page_title: "AWS: aws_sfn_state_machine"
description: |-
  Provides a Step Function State Machine resource.
---

# Resource: aws_sfn_state_machine

Provides a Step Function State Machine resource

## Example Usage
### Basic (Standard Workflow)

```hcl
# ...

resource "aws_sfn_state_machine" "sfn_state_machine" {
  name     = "my-state-machine"
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda.arn}",
      "End": true
    }
  }
}
EOF
}
```

### Basic (Express Workflow)

```hcl
# ...

resource "aws_sfn_state_machine" "sfn_state_machine" {
  name     = "my-state-machine"
  role_arn = aws_iam_role.iam_for_sfn.arn
  type     = "EXPRESS"

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda.arn}",
      "End": true
    }
  }
}
EOF
}
```

### Logging

~> *NOTE:* See the [AWS Step Functions Developer Guide](https://docs.aws.amazon.com/step-functions/latest/dg/welcome.html) for more information about enabling Step Function logging.

```hcl
# ...

resource "aws_sfn_state_machine" "sfn_state_machine" {
  name     = "my-state-machine"
  role_arn = aws_iam_role.iam_for_sfn.arn

  definition = <<EOF
{
  "Comment": "A Hello World example of the Amazon States Language using an AWS Lambda Function",
  "StartAt": "HelloWorld",
  "States": {
    "HelloWorld": {
      "Type": "Task",
      "Resource": "${aws_lambda_function.lambda.arn}",
      "End": true
    }
  }
}
EOF

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.log_group_for_sfn.arn}:*"
    include_execution_data = true
    level                  = "ERROR"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the state machine.
* `definition` - (Required) The Amazon States Language definition of the state machine.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role to use for this state machine.
* `tags` - (Optional) Key-value map of resource tags
* `logging_configuration` - (Optional) Defines what execution history events are logged and where they are logged. The `logging_configuration` parameter is only valid when `type` is set to `EXPRESS`. Defaults to `OFF`. For more information see [Logging Express Workflows](https://docs.aws.amazon.com/step-functions/latest/dg/cw-logs.html) and [Log Levels](https://docs.aws.amazon.com/step-functions/latest/dg/cloudwatch-log-level.html) in the AWS Step Functions User Guide.
* `type` - (Optional) Determines whether a Standard or Express state machine is created. The default is STANDARD. You cannot update the type of a state machine once it has been created. Valid Values: STANDARD | EXPRESS

### `logging_configuration` Configuration Block

* `log_destination` - (Optional) Amazon Resource Name (ARN) of CloudWatch log group. Make sure the State Machine does have the right IAM Policies for Logging. The ARN must end with `:*`
* `include_execution_data` - (Optional) Determines whether execution data is included in your log. When set to FALSE, data is excluded.
* `level` - (Optional) Defines which category of execution history events are logged. Valid Values: ALL | ERROR | FATAL | OFF

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the state machine.
* `creation_date` - The date the state machine was created.
* `status` - The current status of the state machine. Either "ACTIVE" or "DELETING".
* `arn` - The ARN of the state machine.

## Import

State Machines can be imported using the `arn`, e.g.

```
$ terraform import aws_sfn_state_machine.foo arn:aws:states:eu-west-1:123456789098:stateMachine:bar
```
