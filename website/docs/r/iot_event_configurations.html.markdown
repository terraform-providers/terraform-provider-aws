---
layout: "aws"
page_title: "AWS: aws_iot_event_configurations"
description: |-
    Creates and manages an AWS IoT Thing.
---

# Resource: aws_iot_event_configurations

Manages an AWS Event Configurations.

## Example Usage

```hcl
resource "aws_iot_event_configurations" "example" {
    values = {
		"THING" = true,
		"THING_GROUP" = false,
		"THING_TYPE" = false,
		"THING_GROUP_MEMBERSHIP" = false,
		"THING_GROUP_HIERARCHY" = false,
		"THING_TYPE_ASSOCIATION" = false,
		"JOB" = false,
		"JOB_EXECUTION" = false,
		"POLICY" = false,
		"CERTIFICATE" = true,
		"CA_CERTIFICATE" = false,
    }
}
```

## Argument Reference

* `values` - (Optional) Map. The new event configuration values. You can use only these strings as keys: THING_GROUP_HIERARCHY, THING_GROUP_MEMBERSHIP, THING_TYPE, THING_TYPE_ASSOCIATION, THING_GROUP, THING, POLICY, CA_CERTIFICATE, JOB_EXECUTION, CERTIFICATE, JOB. Use boolean for values of mapping.


## Import

IOT Event Configurations can be imported using the name, e.g.

```
$ terraform import aws_iot_event_configurations.example event-configurations
```
