---
layout: "aws"
page_title: "AWS: aws_mq_broker"
sidebar_current: "docs-aws-resource-mq-broker"
description: |-
  Provides an MQ Broker Resource
---

# aws_mq_broker

Provides an MQ Broker Resource. This resources also manages users for the broker.

For more information on Amazon MQ, see [Amazon MQ documentation](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/welcome.html).

Changes to an MQ Broker can occur when you change a
parameter, such as `configuration` or `user`, and are reflected in the next maintenance
window. Because of this, Terraform may report a difference in its planning
phase because a modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change immediately
(see documentation below).

~> **Note:** using `apply_immediately` can result in a
brief downtime as the broker reboots.

~> **Note:** All arguments including the username and password will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
resource "aws_mq_broker" "example" {
  broker_name = "example"

  configuration {
    id       = "${aws_mq_configuration.test.id}"
    revision = "${aws_mq_configuration.test.latest_revision}"
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = ["${aws_security_group.test.id}"]

  user {
    username = "ExampleUser"
    password = "MindTheGap"
  }
}
```

## Argument Reference

The following arguments are supported:

* `apply_immediately` - (Optional) Specifies whether any broker modifications
  are applied immediately, or during the next maintenance window. Default is `false`.
* `auto_minor_version_upgrade` - (Optional) Enables automatic upgrades to new minor versions for brokers, as Apache releases the versions.
* `broker_name` - (Required) The name of the broker.
* `configuration` - (Optional) Configuration of the broker. See below.
* `deployment_mode` - (Optional) The deployment mode of the broker. Supported: `SINGLE_INSTANCE` and `ACTIVE_STANDBY_MULTI_AZ`. Defaults to `SINGLE_INSTANCE`.
* `engine_type` - (Required) The type of broker engine. Currently, Amazon MQ supports only `ActiveMQ`.
* `engine_version` - (Required) The version of the broker engine. Currently, Amazon MQ supports only `5.15.0` or `5.15.6`.
* `host_instance_type` - (Required) The broker's instance type. e.g. `mq.t2.micro` or `mq.m4.large`
* `publicly_accessible` - (Optional) Whether to enable connections from applications outside of the VPC that hosts the broker's subnets.
* `security_groups` - (Required) The list of security group IDs assigned to the broker.
* `subnet_ids` - (Optional) The list of subnet IDs in which to launch the broker. A `SINGLE_INSTANCE` deployment requires one subnet. An `ACTIVE_STANDBY_MULTI_AZ` deployment requires two subnets.
* `maintenance_window_start_time` - (Optional) Maintenance window start time. See below.
* `logs` - (Optional) Logging configuration of the broker. See below.
* `user` - (Optional) The list of all ActiveMQ usernames for the specified broker. See below.
* `tags` - (Optional) A mapping of tags to assign to the resource.

### Nested Fields

#### `configuration`

* `id` - (Optional) The Configuration ID.
* `revision` - (Optional) Revision of the Configuration.

#### `maintenance_window_start_time`

* `day_of_week` - (Required) The day of the week. e.g. `MONDAY`, `TUESDAY`, or `WEDNESDAY`
* `time_of_day` - (Required) The time, in 24-hour format. e.g. `02:00`
* `time_zone` - (Required) The time zone, UTC by default, in either the Country/City format, or the UTC offset format. e.g. `CET`

~> **NOTE:** AWS currently does not support updating the maintenance window beyond resource creation.

### `logs`

* `general` - (Optional) Enables general logging via CloudWatch. Defaults to `false`.
* `audit` - (Optional) Enables audit logging. User management action made using JMX or the ActiveMQ Web Console is logged. Defaults to `false`.

#### `user`

* `console_access` - (Optional) Whether to enable access to the [ActiveMQ Web Console](http://activemq.apache.org/web-console.html) for the user.
* `groups` - (Optional) The list of groups (20 maximum) to which the ActiveMQ user belongs.
* `password` - (Required) The password of the user. It must be 12 to 250 characters long, at least 4 unique characters, and must not contain commas.
* `username` - (Required) The username of the user.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID that Amazon MQ generates for the broker.
* `arn` - The ARN of the broker.
* `instances` - A list of information about allocated brokers (both active & standby).
  * `instances.0.console_url` - The URL of the broker's [ActiveMQ Web Console](http://activemq.apache.org/web-console.html).
  * `instances.0.ip_address` - The IP Address of the broker.
  * `instances.0.endpoints` - The broker's wire-level protocol endpoints in the following order & format referenceable e.g. as `instances.0.endpoints.0` (SSL):
     * `ssl://broker-id.mq.us-west-2.amazonaws.com:61617`
     * `amqp+ssl://broker-id.mq.us-west-2.amazonaws.com:5671`
     * `stomp+ssl://broker-id.mq.us-west-2.amazonaws.com:61614`
     * `mqtt+ssl://broker-id.mq.us-west-2.amazonaws.com:8883`
     * `wss://broker-id.mq.us-west-2.amazonaws.com:61619`

## Import

MQ Broker is currently not importable.
