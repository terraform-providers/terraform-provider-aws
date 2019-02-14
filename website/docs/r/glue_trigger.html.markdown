---
layout: "aws"
page_title: "AWS: aws_glue_trigger"
sidebar_current: "docs-aws-resource-glue-trigger"
description: |-
  Manages a Glue Trigger resource.
---

# aws_glue_trigger

Manages a Glue Trigger resource.

## Example Usage

### Conditional Trigger

```hcl
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    job_name = "${aws_glue_job.example1.name}"
  }

  predicate {
    conditions {
      job_name = "${aws_glue_job.example2.name}"
      state    = "SUCCEEDED"
    }
  }
}
```

### On-Demand Trigger

```hcl
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "ON_DEMAND"

  actions {
    job_name = "${aws_glue_job.example.name}"
  }
}
```

### Scheduled Trigger

```hcl
resource "aws_glue_trigger" "example" {
  name     = "example"
  schedule = "cron(15 12 * * ? *)"
  type     = "SCHEDULED"

  actions {
    job_name = "${aws_glue_job.example.name}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `actions` – (Required) List of actions initiated by this trigger when it fires. Defined below.
* `description` – (Optional) A description of the new trigger.
* `enabled` – (Optional) Start the trigger. Defaults to `true`. Not valid to disable for `ON_DEMAND` type.
* `name` – (Required) The name of the trigger.
* `predicate` – (Optional) A predicate to specify when the new trigger should fire. Required when trigger type is `CONDITIONAL`. Defined below.
* `schedule` – (Optional) A cron expression used to specify the schedule. [Time-Based Schedules for Jobs and Crawlers](https://docs.aws.amazon.com/glue/latest/dg/monitor-data-warehouse-schedule.html)
* `type` – (Required) The type of trigger. Valid values are `CONDITIONAL`, `ON_DEMAND`, and `SCHEDULED`.

### actions Argument Reference

* `arguments` - (Optional) Arguments to be passed to the job. You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes.
* `job_name` - (Required) The name of a job to be executed.
* `timeout` - (Optional) The job run timeout in minutes. It overrides the timeout value of the job.

### predicate Argument Reference

* `conditions` - (Required) A list of the conditions that determine when the trigger will fire. Defined below.
* `logical` - (Optional) How to handle multiple conditions. Defaults to `AND`. Valid values are `AND` or `ANY`.

#### conditions Argument Reference

* `job_name` - (Required) The name of the job to watch.
* `logical_operator` - (Optional) A logical operator. Defaults to `EQUALS`.
* `state` - (Required) The condition state. Currently, the values supported are `SUCCEEDED`, `STOPPED`, `TIMEOUT` and `FAILED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Trigger name

## Timeouts

`aws_glue_trigger` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

- `create` - (Default `5m`) How long to wait for a trigger to be created.
- `delete` - (Default `5m`) How long to wait for a trigger to be deleted.

## Import

Glue Triggers can be imported using `name`, e.g.

```
$ terraform import aws_glue_trigger.MyTrigger MyTrigger
```
