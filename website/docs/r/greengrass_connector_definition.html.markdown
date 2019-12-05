---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_connector_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Connector Definition
---

# Resource: aws_greengrass_connector_definition

## Example Usage

```hcl
resource "aws_greengrass_connector_definition" "definition" {
  name = "test_connector_definition"

  connector_arn = "arn:aws:greengrass:eu-west-1::/connectors/RaspberryPiGPIO/versions/5"
  id = "connector_id"
  parameters = {
    "key" = "value",
  }

}
```

## Argument Reference
* `name` - (Required) The name of the connector definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the connector definition.
* `connector_definition_version` - (Optional) Object. Information about a connector definition version.

The `connector_definition_version` object has such arguments.
* `connector` - (Optional) List of Object. A list of references to connectors in this version, with their corresponding configuration settings.

The `connector` object has such arguments:
* `connector_arn` - (Required). The ARN of the connector.
* `id` - (Required) A descriptive or arbitrary ID for the connector. This value must be unique within the connector definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `parameters` - (Optional) Map. The parameters or configuration that the connector uses.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the group
* `connector_definition_version.arn` - The ARN of latest connector definition version

## Environment variables
If you use `connector_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import

IoT Greengrass Connector Definition can be imported using the `id`, e.g.

```
$ terraform import aws_greengrass_connector_definition.definition <connector_definition_id>
```
