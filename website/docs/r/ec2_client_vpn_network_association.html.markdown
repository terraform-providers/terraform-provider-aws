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

### Using default security group

```hcl
resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  subnet_id              = aws_subnet.example.id
}
```

### Using custom security groups

```hcl
resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  subnet_id              = aws_subnet.example.id
  security_groups        = [aws_security_group.example1.id, aws_security_group.example2.id]
}
```

## Argument Reference

The following arguments are supported:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `subnet_id` - (Required) The ID of the subnet to associate with the Client VPN endpoint.
* `security_groups` - (Optional) A list of up to five custom security groups to apply to the target network. If not specified, the VPC's default security group is assigned.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the target network association.
* `association_id` - The unique ID of the target network association.
* `security_groups` - The IDs of the security groups applied to the target network association.
* `status` - The current state of the target network association.
* `vpc_id` - The ID of the VPC in which the target subnet is located.

## Import

AWS Client VPN network associations can be imported using the endpoint ID and the association ID. Values are separated by a `,`.

```
$ terraform import aws_ec2_client_vpn_network_association.example cvpn-endpoint-0ac3a1abbccddd666,vpn-assoc-0b8db902465d069ad
```
