---
subcategory: "MQ"
layout: "aws"
page_title: "AWS: aws_mq_broker_instance_type_offerings"
description: |-
  Provides a list of availability zones that can deploy an MQ Broker with specified criteria.
---

# Data Source: aws_mq_broker_instance_type_offerings

Provides a list of availability zones that can deploy an MQ Broker with specified criteria. Availability zones
are based on the current active region. If no availability zones can support the criteria given,
an error will be thrown.

## Example Usage

```terraform
data "aws_mq_broker_instance_type_offerings" "test" {
  storage_type       = "EBS"
  engine_type        = "RABBITMQ"
  host_instance_type = "mq.m5.large"
}
```

## Argument Reference

The following arguments are supported:

* `storage_type` - (Optional) Storage type. Valid values are `EBS` and `EFS`
* `engine_type` - (Optional) MQ Broker engine type. Valid values are `RABBITMQ` and `ACTIVEMQ`
* `host_instance_type` - (Optional) The size of the instance

## Attributes Reference

The following attributes are exported:

* `availability_zones` - List of availability zones that can deploy an MQ Broker with specified criteria