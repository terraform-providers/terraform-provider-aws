---
subcategory: "Managed Streaming for Kafka (MSK)"
layout: "aws"
page_title: "AWS: aws_msk_configuration"
description: |-
  Get information on an Amazon MSK Configuration
---

# Data Source: aws_msk_configuration

Get information on an Amazon MSK Configuration.

## Example Usage

```hcl
data "aws_msk_configuration" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the configuration.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the configuration.
* `latest_revision` - Latest revision of the configuration.
* `description` - Description of the configuration.
* `kafka_versions` - List of Apache Kafka versions which can use this configuration.
* `server_properties` - Contents of the server.properties file.
