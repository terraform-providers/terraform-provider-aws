---
subcategory: "Batch"
layout: "aws"
page_title: "AWS: aws_batch_compute_environment"
description: |-
    Provides details about a batch compute environment
---

# Data Source: aws_batch_compute_environment

The Batch Compute Environment data source allows access to details of a specific
compute environment within AWS Batch.

## Example Usage

```hcl
data "aws_batch_compute_environment" "batch-mongo" {
  compute_environment_name = "batch-mongo-production"
}
```

## Argument Reference

The following arguments are supported:

* `compute_environment_name` - (Required) The name of the Batch Compute Environment

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the compute environment.
* `ecs_cluster_arn` - The ARN of the underlying Amazon ECS cluster used by the compute environment.
* `service_role` - The ARN of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.
* `type` - The type of the compute environment (for example, `MANAGED` or `UNMANAGED`).
* `status` - The current status of the compute environment (for example, `CREATING` or `VALID`).
* `status_reason` - A short, human-readable string to provide additional details about the current status of the compute environment.
* `state` - The state of the compute environment (for example, `ENABLED` or `DISABLED`). If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues.
* `tags` - Key-value map of resource tags
