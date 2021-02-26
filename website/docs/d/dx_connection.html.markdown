---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection"
description: |-
  Retrieve information about a Direct Connect Connection
---

# Data Source: aws_dx_connection

Retrieve information about a Direct Connect Connection.

## Example Usage

```hcl
data "aws_dx_connection" "example" {
  connection_id = "dxcon-12345678"
}
```

## Argument Reference

* `connection_id` - (Required) The ID of the connection to retrieve.

## Attributes Reference

* `arn` - The ARN of the connection.
* `aws_device` - The Direct Connect endpoint on which the physical connection terminates.
* `bandwidth` - The bandwidth of the connection.
* `has_logical_redundancy` - Whether the connection supports a secondary BGP peer in the same address family
* `jumbo_frame_capable` - Boolean value representing if jumbo frames are enabled for this connection.
* `lag_id` - The ID of the LAG.
* `location` - The AWS DirectConnect location where the connection is located.
* `name` - The name of the connection.
* `owner_account_id` - AWS Account ID of the connection.
* `partner_name` - The name of the AWS Direct Connect service provider associated with the connection.
* `provider_name` - The name of the service provider associated with the connection.
* `state` - The state of the connection.
* `tags` - Key-value tags for the Direct Connect Connection.
* `vlan` - The ID of the VLAN.
