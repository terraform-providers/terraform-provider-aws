---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_job"
description: |-
  Provides an Glue Job resource.
---

# Resource: aws_glue_job

Provides a Glue Job resource.

-> Glue functionality, such as monitoring and logging of jobs, is typically managed with the `default_arguments` argument. See the [Special Parameters Used by AWS Glue](https://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-etl-glue-arguments.html) topic in the Glue developer guide for additional information.

## Example Usage

### Python Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.py"
  }
}
```

### Scala Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.scala"
  }

  default_arguments = {
    "--job-language" = "scala"
  }
}
```

### Enabling CloudWatch Logs and Metrics

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name              = "example"
  retention_in_days = 14
}

resource "aws_glue_job" "example" {
  # ... other configuration ...

  default_arguments = {
    # ... potentially other arguments ...
    "--continuous-log-logGroup"          = aws_cloudwatch_log_group.example.name
    "--enable-continuous-cloudwatch-log" = "true"
    "--enable-continuous-log-filter"     = "true"
    "--enable-metrics"                   = ""
  }
}
```

## Argument Reference

The following arguments are supported:

* `command` – (Required) The command of the job. Defined below.
* `connections` – (Optional) The list of connections used for this job.
* `default_arguments` – (Optional) The map of default arguments for this job. You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes. For information about how to specify and consume your own Job arguments, see the [Calling AWS Glue APIs in Python](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-calling.html) topic in the developer guide. For information about the key-value pairs that AWS Glue consumes to set up your job, see the [Special Parameters Used by AWS Glue](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-glue-arguments.html) topic in the developer guide.
* `non_overridable_arguments` – (Optional) Non-overridable arguments for this job, specified as name-value pairs.
* `description` – (Optional) Description of the job.
* `execution_property` – (Optional) Execution property of the job. Defined below.
* `glue_version` - (Optional) The version of glue to use, for example "1.0". For information about available versions, see the [AWS Glue Release Notes](https://docs.aws.amazon.com/glue/latest/dg/release-notes.html).
* `max_capacity` – (Optional) The maximum number of AWS Glue data processing units (DPUs) that can be allocated when this job runs. `Required` when `pythonshell` is set, accept either `0.0625` or `1.0`. Use `number_of_workers` and `worker_type` arguments instead with `glue_version` `2.0` and above.
* `max_retries` – (Optional) The maximum number of times to retry this job if it fails.
* `name` – (Required) The name you assign to this job. It must be unique in your account.
* `notification_property` - (Optional) Notification property of the job. Defined below.
* `role_arn` – (Required) The ARN of the IAM role associated with this job.
* `tags` - (Optional) Key-value map of resource tags
* `timeout` – (Optional) The job timeout in minutes. The default is 2880 minutes (48 hours).
* `security_configuration` - (Optional) The name of the Security Configuration to be associated with the job.
* `worker_type` - (Optional) The type of predefined worker that is allocated when a job runs. Accepts a value of Standard, G.1X, or G.2X.
* `number_of_workers` - (Optional) The number of workers of a defined workerType that are allocated when a job runs.

### command Argument Reference

* `name` - (Optional) The name of the job command. Defaults to `glueetl`. Use `pythonshell` for Python Shell Job Type, `max_capacity` needs to be set if `pythonshell` is chosen.
* `script_location` - (Required) Specifies the S3 path to a script that executes a job.
* `python_version` - (Optional) The Python version being used to execute a Python shell job. Allowed values are 2 or 3.

### execution_property Argument Reference

* `max_concurrent_runs` - (Optional) The maximum number of concurrent runs allowed for a job. The default is 1.

### notification_property Argument Reference

* `notify_delay_after` - (Optional) After a job run starts, the number of minutes to wait before sending a job run delay notification.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Glue Job
* `id` - Job name

## Import

Glue Jobs can be imported using `name`, e.g.

```
$ terraform import aws_glue_job.MyJob MyJob
```
