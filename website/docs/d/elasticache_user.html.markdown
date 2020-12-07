---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user"
description: |-
  Get information on an ElastiCache User resource.
---

# Data Source: aws_elasticache_replication_group

Use this data source to get information about an Elasticache User.

## Example Usage

```hcl
data "aws_elasticache_user" "bar" {
  user_id = "example"
}
```

## Argument Reference

The following arguments are supported:

* `user_id` – (Required) The identifier for the user.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `user_id` - The identifier for the user.
* `user_name` - The user name of the user.
* `access_string` - A string for what access a user possesses within the associated ElastiCache replication groups or clusters.
