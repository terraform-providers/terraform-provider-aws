---
subcategory: "EMR Containers"
layout: "aws"
page_title: "AWS: aws_emrcontainers_virtual_cluster"
description: |-
  Retrieve information about an EMR Containers (EMR on EKS) Virtual Cluster
---

# Data Source: aws_emrcontainers_virtual_cluster

Retrieve information about an EMR Containers (EMR on EKS) Virtual Cluster.

## Example Usage

```terraform
data "aws_emrcontainers_virtual_cluster" "example" {
  id = "example id"
}

output "name" {
  value = data.aws_emrcontainers_virtual_cluster.example.name
}

output "arn" {
  value = data.aws_emrcontainers_virtual_cluster.example.arn
}
```

## Argument Reference

* `id` - (Required) The ID of the cluster.

## Attributes Reference

* `id` - The ID of the cluster.
* `name` - The name of the cluster.
* `arn` - The Amazon Resource Name (ARN) of the cluster.
* `container_provider` - Nested attribute containing information about the underlying container provider (EKS cluster) for your EMR Containers cluster.
    * `id` - The name of the container provider that is running your EMR Containers cluster
    * `info` - Nested list containing information about the configuration of the container provider
      * `eks_info` - Nested list containing EKS-specific information about the cluster where the EMR Containers cluster is running
        * `namespace` - The namespace where the EMR Containers cluster is running
    * `type` - The type of the container provider
* `created_at` - The Unix epoch time stamp in seconds for when the cluster was created.
* `state` - The status of the EKS cluster. One of `RUNNING`, `TERMINATING`, `TERMINATED`, `ARRESTED`.
