---
layout: "aws"
page_title: "AWS: batch"
sidebar_current: "docs-aws-resource-batch-job-queue"
description: |-
  Provides a Batch Job Queue resource.
---

# aws_batch_job_queue

Provides a Batch Job Queue resource.

## Example Usage

```hcl
resource "aws_batch_job_queue" "test_queue" {
  name                 = "tf-test-batch-job-queue"
  state                = "ENABLED"
  priority             = 1
  compute_environments = ["${aws_batch_compute_environment.test_environment_1.arn}", "${aws_batch_compute_environment.test_environment_2.arn}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the job queue.
* `compute_environments` - (Required) Specifies the set of compute environments
    mapped to a job queue and their order.  The position of the compute environments
    in the list will dictate the order. You can associate up to 3 compute environments
    with a job queue.
* `priority` - (Required) The priority of the job queue. Job queues with a higher priority
    are evaluated first when associated with the same compute environment.
* `state` - (Required) The state of the job queue. Must be one of: `ENABLED` or `DISABLED`

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of the job queue.
