---
subcategory: "Elastic Map Reduce (EMR)"
layout: "aws"
page_title: "AWS: aws_emr_managed_scaling_policy"
description: |-
  Provides a resource for EMR Managed Scaling policy
---

# Resource: aws_emr_managed_scaling_policy

Provides a Managed Scaling policy for EMR Cluster. With Amazon EMR versions 5.30.0 and later (except for Amazon EMR 6.0.0), you can enable EMR managed scaling to automatically increase or decrease the number of instances or units in your cluster based on workload. See [Using EMR Managed Scaling in Amazon EMR](https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-managed-scaling.html) for more information.

## Example Usage

```hcl
resource "aws_emr_cluster" "sample" {
  name          = "emr-sample-cluster"
  release_label = "emr-5.30.0"

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "c4.large"
  }
  # skip ...
}

resource "aws_emr_managed_scaling_policy" "samplepolicy" {
  cluster_id = aws_emr_cluster.sample.id
  compute_limits {
    unit_type                       = "Instances"
    minimum_capacity_units          = 2
    maximum_capacity_units          = 10
    maximum_ondemand_capacity_units = 2
    maximum_core_capacity_units     = 10
  }
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The id of the EMR cluster
* `compute_limits` - (Required) Configuration block with compute limit settings. Described below.

### compute_limits

* `unit_type` - (Required) The unit type used for specifying a managed scaling policy. Valid Values: `InstanceFleetUnits` | `Instances` | `VCPU`
* `minimum_capacity_units` - (Required) The lower boundary of EC2 units. It is measured through VCPU cores or instances for instance groups and measured through units for instance fleets. Managed scaling activities are not allowed beyond this boundary. The limit only applies to the core and task nodes. The master node cannot be scaled after initial configuration.
* `maximum_capacity_units` - (Required) The upper boundary of EC2 units. It is measured through VCPU cores or instances for instance groups and measured through units for instance fleets. Managed scaling activities are not allowed beyond this boundary. The limit only applies to the core and task nodes. The master node cannot be scaled after initial configuration.
* `maximum_ondemand_capacity_units` - (Optional) The upper boundary of On-Demand EC2 units. It is measured through VCPU cores or instances for instance groups and measured through units for instance fleets. The On-Demand units are not allowed to scale beyond this boundary. The parameter is used to split capacity allocation between On-Demand and Spot instances.
* `maximum_core_capacity_units` - (Optional) The upper boundary of EC2 units for core node type in a cluster. It is measured through VCPU cores or instances for instance groups and measured through units for instance fleets. The core units are not allowed to scale beyond this boundary. The parameter is used to split capacity allocation between core and task nodes.

## Import

EMR Managed Scaling Policies can be imported via the EMR Cluster identifier, e.g.

```console
$ terraform import aws_emr_managed_scaling_policy.example j-123456ABCDEF
```
