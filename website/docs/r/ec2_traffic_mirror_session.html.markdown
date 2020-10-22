---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_traffic_mirror_session"
description: |-
  Provides a Traffic mirror session
---

# Resource: aws_ec2_traffic_mirror_session

Provides an Traffic mirror session.  
Read [limits and considerations](https://docs.aws.amazon.com/vpc/latest/mirroring/traffic-mirroring-considerations.html) for traffic mirroring

## Example Usage

To create a basic traffic mirror session

```hcl
resource "aws_ec2_traffic_mirror_filter" "filter" {
  description      = "traffic mirror filter - terraform example"
  network_services = ["amazon-dns"]
}

resource "aws_ec2_traffic_mirror_target" "target" {
  network_load_balancer_arn = aws_lb.lb.arn
}

resource "aws_ec2_traffic_mirror_session" "session" {
  description              = "traffic mirror session - terraform example"
  network_interface_id     = aws_instance.test.primary_network_interface_id
  traffic_mirror_filter_id = aws_ec2_traffic_mirror_filter.filter.id
  traffic_mirror_target_id = aws_ec2_traffic_mirror_target.target.id
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A description of the traffic mirror session.
* `network_interface_id` - (Required, Forces new) ID of the source network interface. Not all network interfaces are eligible as mirror sources. On EC2 instances only nitro based instances support mirroring.
* `traffic_mirror_filter_id`  - (Required) ID of the traffic mirror filter to be used
* `traffic_mirror_target_id` - (Required) ID of the traffic mirror target to be used
* `packet_length` - (Optional) The number of bytes in each packet to mirror. These are bytes after the VXLAN header. Do not specify this parameter when you want to mirror the entire packet. To mirror a subset of the packet, set this to the length (in bytes) that you want to mirror.
* `session_number` - (Required) - The session number determines the order in which sessions are evaluated when an interface is used by multiple sessions. The first session with a matching filter is the one that mirrors the packets.
* `virtual_network_id` - (Optional) - The VXLAN ID for the Traffic Mirror session. For more information about the VXLAN protocol, see RFC 7348. If you do not specify a VirtualNetworkId, an account-wide unique id is chosen at random.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the traffic mirror session.
* `id` - The name of the session.

## Import

Traffic mirror sessions can be imported using the `id`, e.g.

```
$ terraform import aws_ec2_traffic_mirror_session.session tms-0d8aa3ca35897b82e
```
