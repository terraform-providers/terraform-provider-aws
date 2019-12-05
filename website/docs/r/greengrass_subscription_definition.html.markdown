---
subcategory: "Greengrass"
layout: "aws"
page_title: "AWS: aws_greengrass_subscription_definition"
description: |-
    Creates and manages an AWS IoT Greengrass Subscription Definition
---

# Resource: aws_greengrass_subscription_definition

## Example Usage

```hcl
resource "aws_greengrass_subscription_definition" "test" {
	name = "subscription_definition_%[1]s"
	subscription_definition_version {
		subscription {
			id = "test_id"
			subject = "test_subject"
			source = "arn:aws:iot:eu-west-1:111111111111:thing/Source"
			target = "arn:aws:iot:eu-west-1:222222222222:thing/Target"	
		}
	}
}
```

## Argument Reference
* `name` - (Required) The name of the subscription definition.
* `tags` - (Optional) Map. Map of tags. Metadata that can be used to manage the subscription definition.
* `subscription_definition_version` - (Optional) Object. Information about a subscription definition version.

The `subscription_definition_version` object has such arguments.
* `subscription` - (Optional) List of Object. A list of subscriptions.

The `subscription` object has such arguments:
* `id` - (Required) String. A descriptive or arbitrary ID for the core. This value must be unique within the core definition version. Max length is 128 characters with pattern [a-zA-Z0-9:_-]+.
* `source` - (Required) String. The source of the subscription. Can be a thing ARN, a Lambda function ARN a connector ARN, 'cloud' (which represents the AWS IoT cloud), or 'GGShadowService'.
* `target` - (Required) String. Where the message is sent to. Can be a thing ARN, a Lambda function ARN, a connector ARN, 'cloud' (which represents the AWS IoT cloud), or 'GGShadowService'.
* `subject` - (Required) String. The MQTT topic used to route the message.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:
* `arn` - The ARN of the group
* `latest_definition_version_arn` - The ARN of latest subscription definition version

## Environment variables
If you use `subscription_definition_version` object you should set `AMZN_CLIENT_TOKEN` as environmental variable.

## Import
IoT Greengrass Subscription Definition can be imported using the `id`, e.g.
```
$ terraform import aws_greengrass_subscription_definition.definition <subscription_definition_id>
``` 