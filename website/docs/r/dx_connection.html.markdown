---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection"
description: |-
  Provides a Connection of Direct Connect.
---

# Resource: aws_dx_connection

Provides a Connection of Direct Connect.

## Example Usage

```hcl
resource "aws_dx_connection" "hoge" {
  name      = "tf-dx-connection"
  bandwidth = "1Gbps"
  location  = "EqDC2"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection.
* `bandwidth` - (Required) The bandwidth of the connection. Valid values for dedicated connections: 1Gbps, 10Gbps. Valid values for hosted connections: 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps, 1Gbps, 2Gbps, 5Gbps and 10Gbps. Case sensitive.
* `location` - (Required) The AWS Direct Connect location where the connection is located. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the connection.
* `arn` - The ARN of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection.
* `has_logical_redundancy` - Indicates whether the connection supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.

## Import

Direct Connect connections can be imported using the `connection id`, e.g.

```
$ terraform import aws_dx_connection.test_connection dxcon-ffre0ec3
```
