---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_hosted_connection"
description: |-
  Provides a hosted connection on the specified interconnect or a link aggregation group (LAG) of interconnects. Intended for use by AWS Direct Connect Partners only.
---

# Resource: aws_dx_hosted_connection

Provides a hosted connection on the specified interconnect or a link aggregation group (LAG) of interconnects. Intended for use by AWS Direct Connect Partners only.

## Example Usage

```hcl
resource "aws_dx_hosted_connection" "hosted" {
  connection_id    = "dxcon-ffabc123"
  bandwidth        = "100Mbps"
  name             = "tf-dx-hosted-connection"
  owner_account_id = "123456789012"
  vlan             = 1
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection.
* `bandwidth` - (Required) The bandwidth of the connection. Valid values for dedicated connections: 1Gbps, 10Gbps. Valid values for hosted connections: 50Mbps, 100Mbps, 200Mbps, 300Mbps, 400Mbps, 500Mbps, 1Gbps, 2Gbps, 5Gbps and 10Gbps. Case sensitive.
* `connection_id` - (Required) The ID of the interconnect or LAG.
* `owner_account_id` - (Required) The ID of the AWS account of the customer for the connection.
* `vlan` - (Required) The dedicated VLAN provisioned to the hosted connection.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the connection.
* `jumbo_frame_capable` - Boolean value representing if jumbo frames have been enabled for this connection.
* `has_logical_redundancy` - Indicates whether the connection supports a secondary BGP peer in the same address family (IPv4/IPv6).
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.
* `state` - The state of the connection. Possible values include: ordering, requested, pending, available, down, deleting, deleted, rejected, unknown. See [AllocateHostedConnection](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_AllocateHostedConnection.html) for a description of each connection state.
* `lag_id` - The ID of the LAG.
* `loa_issue_time` - The time of the most recent call to [DescribeLoa](https://docs.aws.amazon.com/directconnect/latest/APIReference/API_DescribeLoa.html) for this connection.
* `location` - The location of the connection.
* `partner_name` - The name of the AWS Direct Connect service provider associated with the connection.
* `provider_name` - The name of the service provider associated with the connection.
* `region` - The AWS Region where the connection is located.
