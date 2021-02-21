---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_listener"
description: |-
  Provides a Global Accelerator listener.
---

# Resource: aws_globalaccelerator_listener

Provides a Global Accelerator listener.

## Example Usage

```hcl
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "example-bucket"
    flow_logs_s3_prefix = "flow-logs/"
  }
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  client_affinity = "SOURCE_IP"
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 80
  }
}
```

## Argument Reference

The following arguments are supported:

* `accelerator_arn` - (Required) The Amazon Resource Name (ARN) of your accelerator.
* `client_affinity` - (Optional) Direct all requests from a user to the same endpoint. Valid values are `NONE`, `SOURCE_IP`. Default: `NONE`. If `NONE`, Global Accelerator uses the "five-tuple" properties of source IP address, source port, destination IP address, destination port, and protocol to select the hash value. If `SOURCE_IP`, Global Accelerator uses the "two-tuple" properties of source (client) IP address and destination IP address to select the hash value.
* `protocol` - (Optional) The protocol for the connections from clients to the accelerator. Valid values are `TCP`, `UDP`.
* `port_range` - (Optional) The list of port ranges for the connections from clients to the accelerator. Fields documented below.

**port_range** supports the following attributes:

* `from_port` - (Optional) The first port in the range of ports, inclusive.
* `to_port` - (Optional) The last port in the range of ports, inclusive.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the listener.

`aws_globalaccelerator_listener` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `30 minutes`) How long to wait for the Global Accelerator Listener to be created.
* `update` - (Default `30 minutes`) How long to wait for the Global Accelerator Listener to be updated.
* `delete` - (Default `30 minutes`) How long to wait for the Global Accelerator Listener to be deleted.

## Import

Global Accelerator listeners can be imported using the `id`, e.g.

```
$ terraform import aws_globalaccelerator_listener.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/listener/xxxxxxxx
```
