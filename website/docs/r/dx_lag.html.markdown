---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_lag"
description: |-
  Provides a Direct Connect LAG.
---

# Resource: aws_dx_lag

Provides a Direct Connect LAG. Connections can be added to the LAG via the [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html) and [`aws_dx_connection_association`](/docs/providers/aws/r/dx_connection_association.html) resources.

~> *NOTE:* When creating a LAG, Direct Connect requires creating a Connection. Terraform will remove this unmanaged connection during resource creation.

## Example Usage

```hcl
resource "aws_dx_lag" "hoge" {
  name                  = "tf-dx-lag"
  connections_bandwidth = "1Gbps"
  location              = "EqDC2"
  force_destroy         = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the LAG.
* `connections_bandwidth` - (Required) The bandwidth of the individual physical connections bundled by the LAG. Available values: 1Gbps, 10Gbps. Case sensitive.
* `location` - (Required) The AWS Direct Connect location in which the LAG should be allocated. See [DescribeLocations](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLocations.html) for the list of AWS Direct Connect locations. Use `locationCode`.
* `force_destroy` - (Optional, Default:false) A boolean that indicates all connections associated with the LAG should be deleted so that the LAG can be destroyed without error. These objects are *not* recoverable.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the LAG.
* `arn` - The ARN of the LAG.
* `jumbo_frame_capable` -Indicates whether jumbo frames (9001 MTU) are supported.
* `has_logical_redundancy` - Indicates whether the LAG supports a secondary BGP peer in the same address family (IPv4/IPv6).

## Import

Direct Connect LAGs can be imported using the `lag id`, e.g.

```
$ terraform import aws_dx_lag.test_lag dxlag-fgnsp5rq
```
