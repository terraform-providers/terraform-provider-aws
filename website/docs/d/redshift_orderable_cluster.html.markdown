---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_orderable_cluster"
description: |-
  Information about RDS orderable DB instances.
---

# Data Source: aws_redshift_orderable_cluster

Information about Redshift Orderable Clusters and valid parameter combinations.

## Example Usage

```hcl
data "aws_redshift_orderable_cluster" "test" {
  cluster_type         = "multi-node"
  preferred_node_types = ["dc2.large", "ds2.xlarge"]
}
```

## Argument Reference

The following arguments are supported:

* `cluster_type` - (Optional) Reshift Cluster type. e.g. `multi-node` or `single-node`
* `cluster_version` - (Optional) Redshift Cluster version. e.g. `1.0`
* `node_type` - (Optional) Redshift Cluster node type. e.g. `dc2.8xlarge`
* `preferred_node_types` - (Optional) Ordered list of preferred Redshift Cluster node types. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `availability_zones` - List of Availability Zone names where the Redshit Cluster is available.
