---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_maintenance_window_task"
description: |-
  Provides an SSM Maintenance Window Task resource
---

# Resource: aws_ssm_maintenance_window_task

Provides an SSM Maintenance Window Task resource

## Example Usage

### Automation Tasks

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  max_concurrency  = 2
  max_errors       = 1
  priority         = 1
  service_role_arn = aws_iam_role.example.arn
  task_arn         = "AWS-RestartEC2Instance"
  task_type        = "AUTOMATION"
  window_id        = aws_ssm_maintenance_window.example.id

  targets {
    key    = "InstanceIds"
    values = [aws_instance.example.id]
  }

  task_invocation_parameters {
    automation_parameters {
      document_version = "$LATEST"

      parameter {
        name   = "InstanceId"
        values = [aws_instance.example.id]
      }
    }
  }
}
```

### Lambda Tasks

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  max_concurrency  = 2
  max_errors       = 1
  priority         = 1
  service_role_arn = aws_iam_role.example.arn
  task_arn         = aws_lambda_function.example.arn
  task_type        = "LAMBDA"
  window_id        = aws_ssm_maintenance_window.example.id

  targets {
    key    = "InstanceIds"
    values = [aws_instance.example.id]
  }

  task_invocation_parameters {
    lambda_parameters {
      client_context = base64encode("{\"key1\":\"value1\"}")
      payload        = "{\"key1\":\"value1\"}"
    }
  }
}
```

### Run Command Tasks

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  max_concurrency  = 2
  max_errors       = 1
  priority         = 1
  service_role_arn = aws_iam_role.example.arn
  task_arn         = "AWS-RunShellScript"
  task_type        = "RUN_COMMAND"
  window_id        = aws_ssm_maintenance_window.example.id

  targets {
    key    = "InstanceIds"
    values = [aws_instance.example.id]
  }

  task_invocation_parameters {
    run_command_parameters {
      output_s3_bucket     = aws_s3_bucket.example.bucket
      output_s3_key_prefix = "output"
      service_role_arn     = aws_iam_role.example.arn
      timeout_seconds      = 600

      notification_config {
        notification_arn    = aws_sns_topic.example.arn
        notification_events = ["All"]
        notification_type   = "Command"
      }

      parameter {
        name   = "commands"
        values = ["date"]
      }
    }
  }
}
```

### Step Function Tasks

```hcl
resource "aws_ssm_maintenance_window_task" "example" {
  max_concurrency  = 2
  max_errors       = 1
  priority         = 1
  service_role_arn = aws_iam_role.example.arn
  task_arn         = aws_sfn_activity.example.id
  task_type        = "STEP_FUNCTIONS"
  window_id        = aws_ssm_maintenance_window.example.id

  targets {
    key    = "InstanceIds"
    values = [aws_instance.example.id]
  }

  task_invocation_parameters {
    step_functions_parameters {
      input = "{\"key1\":\"value1\"}"
      name  = "example"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `window_id` - (Required) The Id of the maintenance window to register the task with.
* `max_concurrency` - (Required) The maximum number of targets this task can be run for in parallel.
* `max_errors` - (Required) The maximum number of errors allowed before this task stops being scheduled.
* `task_type` - (Required) The type of task being registered. Valid values: `AUTOMATION`, `LAMBDA`, `RUN_COMMAND` or `STEP_FUNCTIONS`.
* `task_arn` - (Required) The ARN of the task to execute.
* `service_role_arn` - (Required) The role that should be assumed when executing the task.
* `name` - (Optional) The name of the maintenance window task.
* `description` - (Optional) The description of the maintenance window task.
* `targets` - (Required) The targets (either instances or window target ids). Instances are specified using Key=InstanceIds,Values=instanceid1,instanceid2. Window target ids are specified using Key=WindowTargetIds,Values=window target id1, window target id2.
* `priority` - (Optional) The priority of the task in the Maintenance Window, the lower the number the higher the priority. Tasks in a Maintenance Window are scheduled in priority order with tasks that have the same priority scheduled in parallel.
* `task_invocation_parameters` - (Optional) Configuration block with parameters for task execution.

`task_invocation_parameters` supports the following:

* `automation_parameters` - (Optional) The parameters for an AUTOMATION task type. Documented below.
* `lambda_parameters` - (Optional) The parameters for a LAMBDA task type. Documented below.
* `run_command_parameters` - (Optional) The parameters for a RUN_COMMAND task type. Documented below.
* `step_functions_parameters` - (Optional) The parameters for a STEP_FUNCTIONS task type. Documented below.

`automation_parameters` supports the following:

* `document_version` - (Optional) The version of an Automation document to use during task execution.
* `parameter` - (Optional) The parameters for the RUN_COMMAND task execution. Documented below.

`lambda_parameters` supports the following:

* `client_context` - (Optional) Pass client-specific information to the Lambda function that you are invoking.
* `payload` - (Optional) JSON to provide to your Lambda function as input.
* `qualifier` - (Optional) Specify a Lambda function version or alias name.

`run_command_parameters` supports the following:

* `comment` - (Optional) Information about the command(s) to execute.
* `document_hash` - (Optional) The SHA-256 or SHA-1 hash created by the system when the document was created. SHA-1 hashes have been deprecated.
* `document_hash_type` - (Optional) SHA-256 or SHA-1. SHA-1 hashes have been deprecated. Valid values: `Sha256` and `Sha1`
* `notification_config` - (Optional) Configurations for sending notifications about command status changes on a per-instance basis. Documented below.
* `output_s3_bucket` - (Optional) The name of the Amazon S3 bucket.
* `output_s3_key_prefix` - (Optional) The Amazon S3 bucket subfolder.
* `parameter` - (Optional) The parameters for the RUN_COMMAND task execution. Documented below.
* `service_role_arn` - (Optional) The IAM service role to assume during task execution.
* `timeout_seconds` - (Optional) If this time is reached and the command has not already started executing, it doesn't run.

`step_functions_parameters` supports the following:

* `input` - (Optional) The inputs for the STEP_FUNCTION task.
* `name` - (Optional) The name of the STEP_FUNCTION task.

`notification_config` supports the following:

* `notification_arn` - (Optional) An Amazon Resource Name (ARN) for a Simple Notification Service (SNS) topic. Run Command pushes notifications about command status changes to this topic.
* `notification_events` - (Optional) The different events for which you can receive notifications. Valid values: `All`, `InProgress`, `Success`, `TimedOut`, `Cancelled`, and `Failed`
* `notification_type` - (Optional) When specified with `Command`, receive notification when the status of a command changes. When specified with `Invocation`, for commands sent to multiple instances, receive notification on a per-instance basis when the status of a command changes. Valid values: `Command` and `Invocation`

`parameter` supports the following:

* `name` - (Required) The parameter name.
* `values` - (Required) The array of strings.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the maintenance window task.

## Import

AWS Maintenance Window Task can be imported using the `window_id` and `window_task_id` separated by `/`.

```sh
$ terraform import aws_ssm_maintenance_window_task.task <window_id>/<window_task_id>
```
