---
subcategory: "EMR Containers"
layout: "aws"
page_title: "AWS: aws_emrcontainers_virtual_cluster"
description: |-
  Manages an EMR Containers (EMR on EKS) Virtual Cluster
---

# Resource: aws_emrcontainers_virtual_cluster

Manages an EMR Containers (EMR on EKS) Virtual Cluster.

## Example Usage

### Basic Usage

```terraform
resource "aws_emrcontainers_virtual_cluster" "example" {
  container_provider {
    id   = aws_eks_cluster.example.name
    type = "EKS"

    info {
      eks_info {
        namespace = "default"
      }
    }
  }

  name = "example"

}
```

## Argument Reference

The following arguments are required:

* `name` – (Required) Name of the cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).
* `container_provider` - (Required) Configuration block for the container provider associated with your cluster.

### container_provider Arguments

* `id` - The name of the container provider that is running your EMR Containers cluster
* `info` - Nested list containing information about the configuration of the container provider
  * `eks_info` - Nested list containing EKS-specific information about the cluster where the EMR Containers cluster is running
    * `namespace` - The namespace where the EMR Containers cluster is running
* `type` - The type of the container provider

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the cluster.
* `arn` - ARN of the cluster.
* `created_at` - The Unix epoch time stamp in seconds for when the cluster was created.
* `state` - The status of the EKS cluster. One of `RUNNING`, `TERMINATING`, `TERMINATED`, `ARRESTED`.

## Import

EKS Clusters can be imported using the `name`, e.g.

```
$ terraform import aws_emrcontainers_virtual_cluster.example
```
