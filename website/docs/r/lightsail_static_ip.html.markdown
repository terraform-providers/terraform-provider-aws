---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_static_ip"
description: |-
  Provides an Lightsail Static IP
---

# Resource: aws_lightsail_static_ip

Allocates a static IP address.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage

```hcl
resource "aws_lightsail_static_ip" "test" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the allocated static IP

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Lightsail static IP
* `ip_address` - The allocated static IP address
* `support_code` - The support code.
