---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_rds_cluster"
description: |-
  Provides an RDS cluster data source.
---

# Data Source: aws_rds_cluster

Provides information about an RDS cluster.

## Example Usage

```hcl
data "aws_rds_cluster" "clusterName" {
  cluster_identifier = "clusterName"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required) The cluster identifier of the RDS cluster.

## Attributes Reference

See the [RDS Cluster Resource](/docs/providers/aws/r/rds_cluster.html) for details on the
returned attributes - they are identical.
