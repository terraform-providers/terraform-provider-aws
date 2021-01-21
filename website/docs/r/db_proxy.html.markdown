---
subcategory: "RDS"
layout: "aws"
page_title: "AWS: aws_db_proxy"
description: |-
  Provides an RDS DB proxy resource.
---

# Resource: aws_db_proxy

Provides an RDS DB proxy resource. For additional information, see the [RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/rds-proxy.html).

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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The identifier for the proxy. This name must be unique for all proxies owned by your AWS account in the specified AWS Region. An identifier must begin with a letter and must contain only ASCII letters, digits, and hyphens; it can't end with a hyphen or contain two consecutive hyphens.
* `auth` - (Required) Configuration block(s) with authorization mechanisms to connect to the associated instances or clusters. Described below.
* `debug_logging` - (Optional) Whether the proxy includes detailed information about SQL statements in its logs. This information helps you to debug issues involving SQL behavior or the performance and scalability of the proxy connections. The debug information includes the text of SQL statements that you submit through the proxy. Thus, only enable this setting when needed for debugging, and only when you have security measures in place to safeguard any sensitive information that appears in the logs.
* `engine_family` - (Required, Forces new resource) The kinds of databases that the proxy can connect to. This value determines which database network protocol the proxy recognizes when it interprets network traffic to and from the database. The engine family applies to MySQL and PostgreSQL for both RDS and Aurora. Valid values are `MYSQL` and `POSTGRESQL`.
* `idle_client_timeout` - (Optional) The number of seconds that a connection to the proxy can be inactive before the proxy disconnects it. You can set this value higher or lower than the connection timeout limit for the associated database.
* `require_tls` - (Optional) A Boolean parameter that specifies whether Transport Layer Security (TLS) encryption is required for connections to the proxy. By enabling this setting, you can enforce encrypted TLS connections to the proxy.
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that the proxy uses to access secrets in AWS Secrets Manager.
* `vpc_security_group_ids` - (Optional) One or more VPC security group IDs to associate with the new proxy.
* `vpc_subnet_ids` - (Required) One or more VPC subnet IDs to associate with the new proxy.
* `tags` - (Optional) A mapping of tags to assign to the resource.

`auth` blocks support the following:

* `auth_scheme` - (Optional) The type of authentication that the proxy uses for connections from the proxy to the underlying database. One of `SECRETS`.
* `description` - (Optional) A user-specified description about the authentication used by a proxy to log in as a specific database user.
* `iam_auth` - (Optional) Whether to require or disallow AWS Identity and Access Management (IAM) authentication for connections to the proxy. One of `DISABLED`, `REQUIRED`.
* `secret_arn` - (Optional) The Amazon Resource Name (ARN) representing the secret that the proxy uses to authenticate to the RDS DB instance or Aurora DB cluster. These secrets are stored within Amazon Secrets Manager.
* `username` - (Optional) The name of the database user to which the proxy connects.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) for the proxy.
* `arn` - The Amazon Resource Name (ARN) for the proxy.
* `endpoint` - The endpoint that you can use to connect to the proxy. You include the endpoint value in the connection string for a database client application.

### Timeouts

`aws_db_proxy` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for creating DB proxies.
- `update` - (Default `30 minutes`) Used for modifying DB proxies.
- `delete` - (Default `60 minutes`) Used for destroying DB proxies.

## Import

DB proxies can be imported using the `name`, e.g.

```
$ terraform import aws_db_proxy.example example
```
