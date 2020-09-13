---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_workflow"
description: |-
  Provides an Glue Workflow resource.
---

# Resource: aws_glue_workflow

Provides a Glue Workflow resource.
The workflow graph (DAG) can be build using the `aws_glue_trigger` resource.
See the example below for creating a graph with four nodes (two triggers and two jobs).

## Example Usage

```hcl
resource "aws_glue_workflow" "example" {
  name = "example"
}

resource "aws_glue_trigger" "example-start" {
  name          = "trigger-start"
  type          = "ON_DEMAND"
  workflow_name = aws_glue_workflow.example.name

  actions {
    job_name = "example-job"
  }
}

resource "aws_glue_trigger" "example-inner" {
  name          = "trigger-inner"
  type          = "CONDITIONAL"
  workflow_name = aws_glue_workflow.example.name

  predicate {
    conditions {
      job_name = "example-job"
      state    = "SUCCEEDED"
    }
  }

  actions {
    job_name = "another-example-job"
  }
}
```

## Argument Reference

The following arguments are supported:

* `default_run_properties` – (Optional) A map of default run properties for this workflow. These properties are passed to all jobs associated to the workflow.
* `description` – (Optional) Description of the workflow.
* `name` – (Required) The name you assign to this workflow.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Workflow name

## Import

Glue Workflows can be imported using `name`, e.g.

```
$ terraform import aws_glue_workflow.MyWorkflow MyWorkflow
```
