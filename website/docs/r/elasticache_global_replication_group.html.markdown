---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_global_replication_group"
description: |-
  Provides an ElastiCache Global Replication Group resource.
---

# Resource: aws_elasticache_global_replication_group

Provides an ElastiCache Global Replication Group resource, which manages replication between two or more Replication Groups in different regions. For more information, see the [ElastiCache User Guide](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Redis-Global-Datastore.html).

## Example Usage

### Global replication group with one secondary replication group

The global replication group depends on the primary group existing. Secondary replication groups depend on the global replication group. Terraform dependency management will handle this transparently using resource value references.

```terraform
resource "aws_elasticache_global_replication_group" "example" {
  global_replication_group_id_suffix = "example"
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  replication_group_id          = "example-primary"
  replication_group_description = "primary replication group"

  engine         = "redis"
  engine_version = "5.0.6"
  node_type      = "cache.m5.large"

  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = aws.other_region

  replication_group_id          = "example-secondary"
  replication_group_description = "secondary replication group"
  global_replication_group_id   = aws_elasticache_global_replication_group.example.global_replication_group_id

  number_cache_clusters = 1
}
```

## Argument Reference

The following arguments are supported:

* `global_replication_group_id_suffix` – (Required) The suffix name of a Global Datastore. If `global_replication_group_id_suffix` is changed, creates a new resource.
* `primary_replication_group_id` – (Required) The ID of the primary cluster that accepts writes and will replicate updates to the secondary cluster. If `primary_replication_group_id` is changed, creates a new resource.
* `global_replication_group_description` – (Optional) A user-created description for the global replication group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the ElastiCache Global Replication Group.
* `arn` - The ARN of the ElastiCache Global Replication Group.
* `actual_engine_version` - The full version number of the cache engine running on the members of this global replication group.
* `at_rest_encryption_enabled` - A flag that indicate whether the encryption at rest is enabled.
* `auth_token_enabled` - A flag that indicate whether AuthToken (password) is enabled.
* `cache_node_type` - The instance class used. See AWS documentation for information on [supported node types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html) and [guidance on selecting node types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/nodes-select-size.html).
* `cluster_enabled` - Indicates whether the Global Datastore is cluster enabled.
* `engine` - The name of the cache engine to be used for the clusters in this global replication group.
* `global_replication_group_id` - The full ID of the global replication group.
* `transit_encryption_enabled` - A flag that indicates whether the encryption in transit is enabled.

## Import

ElastiCache Global Replication Groups can be imported using the `global_replication_group_id`, e.g.

```
$ terraform import aws_elasticache_global_replication_group.my_global_replication_group okuqm-global-replication-group-1
```
