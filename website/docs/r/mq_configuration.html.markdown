---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_configuration"
description: |-
  Provides an MQ configuration Resource
---

# Resource: aws_mq_configuration

Provides an MQ Configuration Resource.

For more information on Amazon MQ, see [Amazon MQ documentation](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/welcome.html).

## Example Usage

```terraform
resource "aws_mq_configuration" "example" {
  description    = "Example Configuration"
  name           = "example"
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}
```

## Argument Reference

The following arguments are required:

* `data` - (Required) Broker configuration in XML format. See [official docs](https://docs.aws.amazon.com/amazon-mq/latest/developer-guide/amazon-mq-broker-configuration-parameters.html) for supported parameters and format of the XML.
* `engine_type` - (Required) Type of broker engine. Valid values are `ActiveMQ` and `RabbitMQ`.
* `engine_version` - (Required) Version of the broker engine.
* `name` - (Required) Name of the configuration.

The following arguments are optional:

* `authentication_strategy` - (Optional) Authentication strategy associated with the configuration. Valid values are `simple` and `ldap`. `ldap` is not supported for `engine_type` `RabbitMQ`.
* `description` - (Optional) Description of the configuration.
* `tags` - (Optional) Map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the configuration.
* `id` - Unique ID that Amazon MQ generates for the configuration.
* `latest_revision` - Latest revision of the configuration.

## Import

MQ Configurations can be imported using the configuration ID, e.g.

```
$ terraform import aws_mq_configuration.example c-0187d1eb-88c8-475a-9b79-16ef5a10c94f
```
