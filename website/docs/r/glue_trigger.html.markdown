---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_trigger"
description: |-
  Manages a Glue Trigger resource.
---

# Resource: aws_glue_trigger

Manages a Glue Trigger resource.

## Example Usage

### Conditional Trigger

```hcl
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    job_name = aws_glue_job.example1.name
  }

  predicate {
    conditions {
      job_name = aws_glue_job.example2.name
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
    job_name = aws_glue_job.example.name
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
    job_name = aws_glue_job.example.name
  }
}
```

### Conditional Trigger with Crawler Action

**Note:** Triggers can have both a crawler action and a crawler condition, just no example provided.

```hcl
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    crawler_name = aws_glue_crawler.example1.name
  }

  predicate {
    conditions {
      job_name = aws_glue_job.example2.name
      state    = "SUCCEEDED"
    }
  }
}
```

### Conditional Trigger with Crawler Condition

**Note:** Triggers can have both a crawler action and a crawler condition, just no example provided.

```hcl
resource "aws_glue_trigger" "example" {
  name = "example"
  type = "CONDITIONAL"

  actions {
    job_name = aws_glue_job.example1.name
  }

  predicate {
    conditions {
      crawler_name = aws_glue_crawler.example2.name
      crawl_state  = "SUCCEEDED"
    }
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
* `tags` - (Optional) Key-value map of resource tags
* `type` – (Required) The type of trigger. Valid values are `CONDITIONAL`, `ON_DEMAND`, and `SCHEDULED`.
* `workflow_name` - (Optional) A workflow to which the trigger should be associated to. Every workflow graph (DAG) needs a starting trigger (`ON_DEMAND` or `SCHEDULED` type) and can contain multiple additional `CONDITIONAL` triggers.

### actions Argument Reference

* `arguments` - (Optional) Arguments to be passed to the job. You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes.
* `crawler_name` - (Optional) The name of the crawler to be executed. Conflicts with `job_name`.
* `job_name` - (Optional) The name of a job to be executed. Conflicts with `crawler_name`.
* `timeout` - (Optional) The job run timeout in minutes. It overrides the timeout value of the job.
* `security_configuration` - (Optional) The name of the Security Configuration structure to be used with this action.
* `notification_property` - (Optional) Specifies configuration properties of a job run notification. see [Notification Property](#notification-property) details below.

#### Notification Property

* `notify_delay_after` - (Optional) After a job run starts, the number of minutes to wait before sending a job run delay notification.

### predicate Argument Reference

* `conditions` - (Required) A list of the conditions that determine when the trigger will fire. Defined below.
* `logical` - (Optional) How to handle multiple conditions. Defaults to `AND`. Valid values are `AND` or `ANY`.

#### conditions Argument Reference

* `job_name` - (Optional) The name of the job to watch. If this is specified, `state` must also be specified. Conflicts with `crawler_name`.
* `state` - (Optional) The condition job state. Currently, the values supported are `SUCCEEDED`, `STOPPED`, `TIMEOUT` and `FAILED`. If this is specified, `job_name` must also be specified. Conflicts with `crawler_state`.
* `crawler_name` - (Optional) The name of the crawler to watch. If this is specified, `crawl_state` must also be specified. Conflicts with `job_name`.
* `crawl_state` - (Optional) The condition crawl state. Currently, the values supported are `RUNNING`, `SUCCEEDED`, `CANCELLED`, and `FAILED`. If this is specified, `crawler_name` must also be specified. Conflicts with `state`.
* `logical_operator` - (Optional) A logical operator. Defaults to `EQUALS`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Glue Trigger
* `id` - Trigger name

## Timeouts

`aws_glue_trigger` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `create` - (Default `5m`) How long to wait for a trigger to be created.
- `delete` - (Default `5m`) How long to wait for a trigger to be deleted.

## Import

Glue Triggers can be imported using `name`, e.g.

```
$ terraform import aws_glue_trigger.MyTrigger MyTrigger
```
