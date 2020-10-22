---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_proxy_target"
description: |-
  Provides an RDS DB proxy target resource.
---

# Resource: aws_db_proxy_target

Provides an RDS DB proxy target resource.

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

resource "aws_db_proxy_target" "example" {
  db_instance_identifier = aws_db_instance.example.id
  db_proxy_name          = aws_db_proxy.example.db_proxy_name
  target_group_name      = aws_db_proxy_default_target_group.example.name
}
```

## Argument Reference

The following arguments are supported:

* `db_proxy_name` - (Required, Forces new resource) The name of the DB proxy.
* `target_group_name` - (Required, Forces new resource) The name of the target group.
* `db_instance_identifier` - (Optional, Forces new resource) DB instance identifier.
* `db_cluster_identifier` - (Optional, Forces new resource) DB cluster identifier.

**NOTE:** Either `db_instance_identifier` or `db_cluster_identifier` should be specified and both should not be specified together

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint` - Hostname for the target RDS DB Instance. Only returned for `RDS_INSTANCE` type.
* `id` - Identifier of  `db_proxy_name`, `target_group_name`, target type (e.g. `RDS_INSTANCE` or `TRACKED_CLUSTER`), and resource identifier separated by forward slashes (`/`).
* `port` - Port for the target RDS DB Instance or Aurora DB Cluster.
* `rds_resource_id` - Identifier representing the DB Instance or DB Cluster target.
* `target_arn` - Amazon Resource Name (ARN) for the DB instance or DB cluster. Currently not returned by the RDS API.
* `tracked_cluster_id` - DB Cluster identifier for the DB Instance target. Not returned unless manually importing an `RDS_INSTANCE` target that is part of a DB Cluster.
* `type` - Type of target. e.g. `RDS_INSTANCE` or `TRACKED_CLUSTER`

## Import

RDS DB Proxy Targets can be imported using the `db_proxy_name`, `target_group_name`, target type (e.g. `RDS_INSTANCE` or `TRACKED_CLUSTER`), and resource identifier separated by forward slashes (`/`), e.g.

Instances:

```
$ terraform import aws_db_proxy_target.example example-proxy/default/RDS_INSTANCE/example-instance
```

Provisioned Clusters:

```
$ terraform import aws_db_proxy_target.example example-proxy/default/TRACKED_CLUSTER/example-cluster
```
