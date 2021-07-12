---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user_group"
description: |-
  Provides an ElastiCache User Group resource.
---

# Resource: aws_elasticache_user_group

Provides an ElastiCache User Group resource.

## Example Usage

### Redis User Group Creation

To create user group for ElastiCache Redis:

```hcl
resource "aws_elasticache_user_group" "example" {
  user_group_id = "tf-user-group-1"
  engine        = "redis"
}
```

## Argument Reference

The following arguments are supported:

* `user_group_id` – (Required) The user group identifier. This parameter is stored as a lowercase string.
* `engine` - (Optional) The name of the cache engine to be used for the clusters in this user group. e.g. `redis`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the ElastiCache User Group.

## Timeouts

`aws_elasticache_user_group` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `10m`) How long to wait for a user group to be created.
* `delete` - (Default `10m`) How long to wait for a user group to be deleted.
* `update` - (Default `10m`) How long to wait for user group settings to be updated. This is also separately used for adding/removing replicas and online resize operation completion, if necessary.

## Import

ElastiCache User Groups can be imported using the `user_group_id`, e.g.

```
$ terraform import aws_elasticache_user_group.my_user_group user-group-1
```
