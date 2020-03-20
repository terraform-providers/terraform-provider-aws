---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_network_association"
description: |-
  Provides network associations for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_network_association

Provides network associations for AWS Client VPN endpoints. For more information on usage, please see the 
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```hcl
resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.example.id}"
  subnet_id              = "${aws_subnet.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `subnet_id` - (Required) The ID of the subnet to associate with the Client VPN endpoint.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the target network association.
* `security_groups` - The IDs of the security groups applied to the target network association.
* `status` - The current state of the target network association.
* `vpc_id` - The ID of the VPC in which the target network (subnet) is located. 
