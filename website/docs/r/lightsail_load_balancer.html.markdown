---
subcategory: "Lightsail"
layout: "aws"
page_title: "AWS: aws_lightsail_load_balancer"
description: |-
  Provides a Lightsail Load Balancer
---

# Resource: aws_lightsail_load_balancer

Creates a Lightsail load balancer resource.

~> **Note:** Lightsail is currently only supported in a limited number of AWS Regions, please see ["Regions and Availability Zones in Amazon Lightsail"](https://lightsail.aws.amazon.com/ls/docs/overview/article/understanding-regions-and-availability-zones-in-amazon-lightsail) for more details

## Example Usage, creating a new load balancer

```hcl
resource "aws_lightsail_load_balancer" "load_balancer_test" {
  name              = "test-load-balancer"
  health_check_path = "/"
  instance_port     = "80"
  tags = {
    foo = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Lightsail load balancer.
* `instance_port` - (Required) The instance port the load balancer will connect.
* `health_check_path` - (Optional) The health check path of the load balancer. Default value "/".
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `arn` - The ARN of the Lightsail load balancer.
* `created_at` - The timestamp when the load balancer was created.
* `dns_name` - The DNS name of the load balancer.
* `id` - The name used for this load balancer.
* `protocol` - The protocol of the load balancer.
* `public_ports` - The public ports of the load balancer.

Lightsail Load Balancers can be imported using their name, e.g.

```
$ terraform import aws_lightsail_load_balancer.website_load_balancer_ 'example-load-balancer'
```
