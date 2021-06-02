---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker"
description: |-
  Provides an MQ Broker Resource
---

# Resource: aws_mq_broker

Provides an Amazon MQ broker resource. This resources also manages users for the broker.

-> For more information on Amazon MQ, see [Amazon MQ documentation](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/welcome.html).

~> **NOTE:** Amazon MQ currently places limits on **RabbitMQ** brokers. For example, a RabbitMQ broker cannot have: instances with an associated IP address of an ENI attached to the broker, an associated LDAP server to authenticate and authorize broker connections, storage type `EFS`, audit logging, or `configuration` blocks. Although this resource allows you to create RabbitMQ users, RabbitMQ users cannot have console access or groups. Also, Amazon MQ does not return information about RabbitMQ users so drift detection is not possible.

~> **NOTE:** Changes to an MQ Broker can occur when you change a parameter, such as `configuration` or `user`, and are reflected in the next maintenance window. Because of this, Terraform may report a difference in its planning phase because a modification has not yet taken place. You can use the `apply_immediately` flag to instruct the service to apply the change immediately (see documentation below). Using `apply_immediately` can result in a brief downtime as the broker reboots.

~> **NOTE:** All arguments including the username and password will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).


## Example Usage

### Basic Example

```terraform
resource "aws_mq_broker" "example" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "ExampleUser"
    password = "MindTheGap"
  }
}
```

### High-throughput Optimized Example

This example shows the use of EBS storage for high-throughput optimized performance.

```terraform
resource "aws_mq_broker" "example" {
  broker_name = "example"

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  storage_type       = "ebs"
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "ExampleUser"
    password = "MindTheGap"
  }
}
```

## Argument Reference

The following arguments are required:

* `broker_name` - (Required) Name of the broker.
* `engine_type` - (Required) Type of broker engine. Valid values are `ActiveMQ` and `RabbitMQ`.
* `engine_version` - (Required) Version of the broker engine. See the [AmazonMQ Broker Engine docs](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/broker-engine.html) for supported versions. For example, `5.15.0`.
* `host_instance_type` - (Required) Broker's instance type. For example, `mq.t3.micro`, `mq.m5.large`.
* `user` - (Required) Configuration block for broker users. For `engine_type` of `RabbitMQ`, Amazon MQ does not return broker users preventing this resource from making user updates and drift detection. Detailed below.

The following arguments are optional:

* `apply_immediately` - (Optional) Specifies whether any broker modifications are applied immediately, or during the next maintenance window. Default is `false`.
* `authentication_strategy` - (Optional) Authentication strategy used to secure the broker. Valid values are `simple` and `ldap`. `ldap` is not supported for `engine_type` `RabbitMQ`.
* `auto_minor_version_upgrade` - (Optional) Whether to automatically upgrade to new minor versions of brokers as Amazon MQ makes releases available.
* `configuration` - (Optional) Configuration block for broker configuration. Applies to `engine_type` of `ActiveMQ` only. Detailed below.
* `deployment_mode` - (Optional) Deployment mode of the broker. Valid values are `SINGLE_INSTANCE`, `ACTIVE_STANDBY_MULTI_AZ`, and `CLUSTER_MULTI_AZ`. Default is `SINGLE_INSTANCE`.
* `encryption_options` - (Optional) Configuration block containing encryption options. Detailed below.
* `ldap_server_metadata` - (Optional) Configuration block for the LDAP server used to authenticate and authorize connections to the broker. Not supported for `engine_type` `RabbitMQ`. Detailed below. (Currently, AWS may not process changes to LDAP server metadata.)
* `logs` - (Optional) Configuration block for the logging configuration of the broker. Detailed below.
* `maintenance_window_start_time` - (Optional) Configuration block for the maintenance window start time. Detailed below.
* `publicly_accessible` - (Optional) Whether to enable connections from applications outside of the VPC that hosts the broker's subnets.
* `security_groups` - (Optional) List of security group IDs assigned to the broker.
* `storage_type` - (Optional) Storage type of the broker. For `engine_type` `ActiveMQ`, the valid values are `efs` and `ebs`, and the AWS-default is `efs`. For `engine_type` `RabbitMQ`, only `ebs` is supported. When using `ebs`, only the `mq.m5` broker instance type family is supported.
* `subnet_ids` - (Optional) List of subnet IDs in which to launch the broker. A `SINGLE_INSTANCE` deployment requires one subnet. An `ACTIVE_STANDBY_MULTI_AZ` deployment requires multiple subnets.
* `tags` - (Optional) Map of tags to assign to the broker. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### configuration

The following arguments are optional:

* `id` - (Optional) The Configuration ID.
* `revision` - (Optional) Revision of the Configuration.

### encryption_options

The following arguments are optional:

* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of Key Management Service (KMS) Customer Master Key (CMK) to use for encryption at rest. Requires setting `use_aws_owned_key` to `false`. To perform drift detection when AWS-managed CMKs or customer-managed CMKs are in use, this value must be configured.
* `use_aws_owned_key` - (Optional) Whether to enable an AWS-owned KMS CMK that is not in your account. Defaults to `true`. Setting to `false` without configuring `kms_key_id` will create an AWS-managed CMK aliased to `aws/mq` in your account.

### ldap_server_metadata

The following arguments are optional:

* `hosts` - (Optional) List of a fully qualified domain name of the LDAP server and an optional failover server.
* `role_base` - (Optional) Fully qualified name of the directory to search for a user’s groups.
* `role_name` - (Optional) Specifies the LDAP attribute that identifies the group name attribute in the object returned from the group membership query.
* `role_search_matching` - (Optional) Search criteria for groups.
* `role_search_subtree` - (Optional) Whether the directory search scope is the entire sub-tree.
* `service_account_password` - (Optional) Service account password.
* `service_account_username` - (Optional) Service account username.
* `user_base` - (Optional) Fully qualified name of the directory where you want to search for users.
* `user_role_name` - (Optional) Specifies the name of the LDAP attribute for the user group membership.
* `user_search_matching` - (Optional) Search criteria for users.
* `user_search_subtree` - (Optional) Whether the directory search scope is the entire sub-tree.

### logs

The following arguments are optional:

* `audit` - (Optional) Enables audit logging. Auditing is only possible for `engine_type` of `ActiveMQ`. User management action made using JMX or the ActiveMQ Web Console is logged. Defaults to `false`.
* `general` - (Optional) Enables general logging via CloudWatch. Defaults to `false`.

### maintenance_window_start_time

The following arguments are required:

* `day_of_week` - (Required) Day of the week, e.g. `MONDAY`, `TUESDAY`, or `WEDNESDAY`.
* `time_of_day` - (Required) Time, in 24-hour format, e.g. `02:00`.
* `time_zone` - (Required) Time zone in either the Country/City format or the UTC offset format, e.g. `CET`.

~> **NOTE:** Amazon MQ currently does not support updating the maintenance window. Changes to the maintenance window start time will force a new broker to be created.

### user

* `console_access` - (Optional) Whether to enable access to the [ActiveMQ Web Console](http://activemq.apache.org/web-console.html) for the user. Applies to `engine_type` of `ActiveMQ` only.
* `groups` - (Optional) List of groups (20 maximum) to which the ActiveMQ user belongs. Applies to `engine_type` of `ActiveMQ` only.
* `password` - (Required) Password of the user. It must be 12 to 250 characters long, at least 4 unique characters, and must not contain commas.
* `username` - (Required) Username of the user.

~> **NOTE:** AWS currently does not support updating RabbitMQ users. Updates to users can only be in the RabbitMQ UI.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the broker.
* `id` - Unique ID that Amazon MQ generates for the broker.
* `instances` - List of information about allocated brokers (both active & standby).
    * `instances.0.console_url` - The URL of the broker's [ActiveMQ Web Console](http://activemq.apache.org/web-console.html).
    * `instances.0.ip_address` - IP Address of the broker.
    * `instances.0.endpoints` - Broker's wire-level protocol endpoints in the following order & format referenceable e.g. as `instances.0.endpoints.0` (SSL):
        * For `ActiveMQ`:
            * `ssl://broker-id.mq.us-west-2.amazonaws.com:61617`
            * `amqp+ssl://broker-id.mq.us-west-2.amazonaws.com:5671`
            * `stomp+ssl://broker-id.mq.us-west-2.amazonaws.com:61614`
            * `mqtt+ssl://broker-id.mq.us-west-2.amazonaws.com:8883`
            * `wss://broker-id.mq.us-west-2.amazonaws.com:61619`
        * For `RabbitMQ`:
            * `amqps://broker-id.mq.us-west-2.amazonaws.com:5671`
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

MQ Brokers can be imported using their broker id, e.g.

```
$ terraform import aws_mq_broker.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
