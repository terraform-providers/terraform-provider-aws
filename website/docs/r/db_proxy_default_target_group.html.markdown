---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_proxy_default_target_group"
description: |-
  Manage an RDS DB proxy default target group resource.
---

# Resource: aws_db_proxy_default_target_group

Provides a resource to manage an RDS DB proxy default target group resource.

The `aws_db_proxy_default_target_group` behaves differently from normal resources, in that Terraform does not _create_ or _destroy_ this resource, since it implicitly exists as part of an RDS DB Proxy. On Terraform resource creation it is automatically imported and on resource destruction, Terraform performs no actions in RDS.

## Example Usage

```hcl
resource "aws_db_proxy" "example" {
  name                   = "example"
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.example.arn
  vpc_security_group_ids = [aws_security_group.example.id]
  vpc_subnet_ids         = [aws_subnet.example.id]

  auth {
    auth_scheme = "SECRETS"
    description = "example"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.example.arn
  }

  tags = {
    Name = "example"
    Key  = "value"
  }
}

resource "aws_db_proxy_default_target_group" "example" {
  db_proxy_name = aws_db_proxy.example.name

  connection_pool_config {
    connection_borrow_timeout    = 120
    init_query                   = "SET x=1, y=2"
    max_connections_percent      = 100
    max_idle_connections_percent = 50
    session_pinning_filters      = ["EXCLUDE_VARIABLE_SETS"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `db_proxy_name` - (Required) Name of the RDS DB Proxy.
* `connection_pool_config` - (Optional) The settings that determine the size and behavior of the connection pool for the target group.

`connection_pool_config` blocks support the following:

* `connection_borrow_timeout` - (Optional) The number of seconds for a proxy to wait for a connection to become available in the connection pool. Only applies when the proxy has opened its maximum number of connections and all connections are busy with client sessions.
* `init_query` - (Optional) One or more SQL statements for the proxy to run when opening each new database connection. Typically used with `SET` statements to make sure that each connection has identical settings such as time zone and character set. This setting is empty by default. For multiple statements, use semicolons as the separator. You can also include multiple variables in a single `SET` statement, such as `SET x=1, y=2`.
* `max_connections_percent` - (Optional) The maximum size of the connection pool for each target in a target group. For Aurora MySQL, it is expressed as a percentage of the max_connections setting for the RDS DB instance or Aurora DB cluster used by the target group.
* `max_idle_connections_percent` - (Optional) Controls how actively the proxy closes idle database connections in the connection pool. A high value enables the proxy to leave a high percentage of idle connections open. A low value causes the proxy to close idle client connections and return the underlying database connections to the connection pool. For Aurora MySQL, it is expressed as a percentage of the max_connections setting for the RDS DB instance or Aurora DB cluster used by the target group.
* `session_pinning_filters` - (Optional) Each item in the list represents a class of SQL operations that normally cause all later statements in a session using a proxy to be pinned to the same underlying database connection. Including an item in the list exempts that class of SQL operations from the pinning behavior. Currently, the only allowed value is `EXCLUDE_VARIABLE_SETS`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Name of the RDS DB Proxy.
* `arn` - The Amazon Resource Name (ARN) representing the target group.
* `name` - The name of the default target group.

### Timeouts

`aws_db_proxy_default_target_group` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `30 minutes`) Timeout for modifying DB proxy target group on creation.
- `update` - (Default `30 minutes`) Timeout for modifying DB proxy target group on update.

## Import

DB proxy default target groups can be imported using the `db_proxy_name`, e.g.

```
$ terraform import aws_db_proxy_default_target_group.example example
```
