---
subcategory: "Chime"
layout: "aws"
page_title: "AWS: aws_chime_voice_connector"
description: |-
  Enables you to connect your phone system to the telephone network at a substantial cost savings by using SIP trunking.
---

# Resource: aws_chime_voice_connector

Enables you to connect your phone system to the telephone network at a substantial cost savings by using SIP trunking.

## Example Usage

```terraform
resource "aws_chime_voice_connector" "test" {
  name               = "connector-test-1"
  require_encryption = true
  aws_region         = "us-east-1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Amazon Chime Voice Connector.
* `require_encryption` - (Required) When enabled, requires encryption for the Amazon Chime Voice Connector.
* `aws_region` - (Optional) The AWS Region in which the Amazon Chime Voice Connector is created. Default value: `us-east-1`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `outbound_host_name` - The outbound host name for the Amazon Chime Voice Connector.

## Import

Configuration Recorder can be imported using the name, e.g.

```
$ terraform import aws_chime_voice_connector.test example
```
