---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_instance_public_ports"
description: |-
  Provides an Lightsail Instance
---

# Resource: aws_lightsail_instance_public_ports

Opens ports for a specific Amazon Lightsail instance, and specifies the IP addresses allowed to connect to the instance through the ports, and the protocol.

-> See [What is Amazon Lightsail?](https://lightsail.aws.amazon.com/ls/docs/getting-started/article/what-is-amazon-lightsail) for more information.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details.

## Example Usage

```terraform
resource "aws_lightsail_instance" "test" {
  name              = "yak_sail"
  availability_zone = data.aws_availability_zones.available.names[0]
  blueprint_id      = "amazon_linux"
  bundle_id         = "nano_1_0"
}

resource "aws_lightsail_instance_public_ports" "test" {
  instance_name = aws_lightsail_instance.test.name

  port_info {
    protocol  = "tcp"
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

The following arguments are required:

* `instance_name` - (Required) Name of the Lightsail Instance.
* `port_info` - (Required) Configuration block with port information. AWS closes all currently open ports that are not included in the `port_info`. Detailed below.

### port_info

The following arguments are required:

* `from_port` - (Required) First port in a range of open ports on an instance.
* `protocol` - (Required) IP protocol name. Valid values are `tcp`, `all`, `udp`, and `icmp`.
* `to_port` - (Required) Last port in a range of open ports on an instance.

The following arguments are optional:

* `cidrs` - (Optional) Set of CIDR blocks.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
