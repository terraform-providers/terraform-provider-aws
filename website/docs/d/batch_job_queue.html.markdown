---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_job_queue"
description: |-
    Provides details about a batch job queue
---

# Data Source: aws_batch_job_queue

The Batch Job Queue data source allows access to details of a specific
job queue within AWS Batch.

## Example Usage

```hcl
data "aws_batch_job_queue" "test-queue" {
  name = "tf-test-batch-job-queue"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the job queue.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the job queue.
* `status` - The current status of the job queue (for example, `CREATING` or `VALID`).
* `status_reason` - A short, human-readable string to provide additional details about the current status
    of the job queue.
* `state` - Describes the ability of the queue to accept new jobs (for example, `ENABLED` or `DISABLED`).
* `tags` - Key-value map of resource tags
* `priority` - The priority of the job queue. Job queues with a higher priority are evaluated first when
    associated with the same compute environment.
* `compute_environment_order` - The compute environments that are attached to the job queue and the order in
    which job placement is preferred. Compute environments are selected for job placement in ascending order.
    * `compute_environment_order.#.order` - The order of the compute environment.
    * `compute_environment_order.#.compute_environment` - The ARN of the compute environment.
