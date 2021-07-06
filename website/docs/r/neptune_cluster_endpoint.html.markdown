---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_endpoint"
description: |-
  Provides an Neptune Cluster Endpoint Resource
---

# Resource: aws_neptune_cluster_endpoint

Provides an Neptune Cluster Endpoint Resource.

## Example Usage

```terraform
resource "aws_neptune_cluster_endpoint" "example" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = "example"
  endpoint_type               = "READER"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required, Forces new resources) The DB cluster identifier of the DB cluster associated with the endpoint.
* `cluster_identifier_endpoint` - (Required, Forces new resources) The identifier of the endpoint.
* `endpoint_type` - (Required) The type of the endpoint. One of: `READER`, `WRITER`, `ANY`.
* `excluded_members` - (Optional) List of DB instance identifiers that aren't part of the custom endpoint group. All other eligible instances are reachable through the custom endpoint. Only relevant if the list of static members is empty.
* `static_members` - (Optional) List of DB instance identifiers that are part of the custom endpoint group.
* `tags` - (Optional) A map of tags to assign to the Neptune cluster. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Neptune Cluster Endpoint Amazon Resource Name (ARN).
* `endpoint` - The DNS address of the endpoint.
* `id` - The Neptune Cluster Endpoint Identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_neptune_cluster_endpoint` can be imported by using the `cluster-identifier:endpoint-identfier`, e.g.

```
$ terraform import aws_neptune_cluster_endpoint.example my-cluster:my-endpoint
```
