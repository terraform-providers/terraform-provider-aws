---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_service"
description: |-
    Provides details about an ecs service
---

# Data Source: aws_ecs_service

The ECS Service data source allows access to details of a specific
Service within a AWS ECS Cluster.

## Example Usage

```hcl
data "aws_ecs_service" "example" {
  service_name = "example"
  cluster_arn  = data.aws_ecs_cluster.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `service_name` - (Required) The name of the ECS Service
* `cluster_arn` - (Required) The arn of the ECS Cluster

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the ECS Service
* `desired_count` - The number of tasks for the ECS Service
* `launch_type` - The launch type for the ECS Service
* `scheduling_strategy` - The scheduling strategy for the ECS Service
* `task_definition` - The family for the latest ACTIVE revision
