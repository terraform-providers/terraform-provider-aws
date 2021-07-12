---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_user"
description: |-
  Provides an ElastiCache User resource.
---

# Resource: aws_elasticache_user

Provides an ElastiCache User resource.

## Example Usage

Here's a basic example of using `aws_elasticache_user` with no passwords configured:

```hcl
resource "aws_elasticache_user" "test" {
  user_id       = "test-user-id"
  user_name     = "test-user-name"
  access_string = "on ~* +@all"
}
```

Here's a more advanced example of using `aws_elasticache_user` with passwords configured:

```hcl
resource "aws_elasticache_user" "test" {
  user_id              = "test-user-id"
  user_name            = "test-user-name"
  access_string        = "on ~* +@all"
  engine               = "redis"
  no_password_required = false
  passwords            = ["password1234567890", "password0123456789"]
}
```

## Argument Reference

The following arguments are supported:

* `user_id` – (Required) User ID of the Elasticache User.
* `user_name` – (Required) User Name of the Elasticache User.
* `access_string` – (Required) List of space-delimited rules which are applied on the Elasticache User.
* `engine` – (Optional) Name of the cache engine to be used.  Valid value for this parameter is `redis`.
* `no_password_required` - (Optional) Whether the ElastiCache User will have passwords. Valid values for this parameters are `true` and `false`. Default is set to `true`.
* `passwords` - (Optional) Set of Passwords configured for the ElastiCache User if `no_password_required` is set to `false`.  This has a minimum of `1` password entry and a maximum of `2` password entries.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `user_id`


## Import

ElastiCache Users can be imported using the `user_id`, e.g.

```
$ terraform import aws_elasticache_user.foo tf-test-user-id
```
