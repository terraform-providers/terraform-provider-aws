---
layout: "aws"
page_title: "AWS: aws_eip"
sidebar_current: "docs-aws-resource-eip"
description: |-
  Provides an Elastic IP resource.
---

# aws_eip

Provides an Elastic IP resource.

~> **Note:** EIP may require IGW to exist prior to association. Use `depends_on` to set an explicit dependency on the IGW.

~> **Note:** Do not use `network_interface` to associate the EIP to `aws_lb` or `aws_nat_gateway` resources. Instead use the `allocation_id` available in those resources to allow AWS to manage the association, otherwise you will see `AuthFailure` errors.

## Example Usage

Single EIP associated with an instance:

```hcl
resource "aws_eip" "lb" {
  instance = "${aws_instance.web.id}"
  vpc      = true
}
```

Multiple EIPs associated with a single network interface:

```hcl
resource "aws_network_interface" "multi-ip" {
  subnet_id   = "${aws_subnet.main.id}"
  private_ips = ["10.0.0.10", "10.0.0.11"]
}

resource "aws_eip" "one" {
  vpc                       = true
  network_interface         = "${aws_network_interface.multi-ip.id}"
  associate_with_private_ip = "10.0.0.10"
}

resource "aws_eip" "two" {
  vpc                       = true
  network_interface         = "${aws_network_interface.multi-ip.id}"
  associate_with_private_ip = "10.0.0.11"
}
```

Attaching an EIP to an Instance with a pre-assigned private ip (VPC Only):

```hcl
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.default.id}"
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_instance" "foo" {
  # us-west-2
  ami           = "ami-5189a661"
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"
}

resource "aws_eip" "bar" {
  vpc = true

  instance                  = "${aws_instance.foo.id}"
  associate_with_private_ip = "10.0.0.12"
  depends_on                = ["aws_internet_gateway.gw"]
}
```

Allocating EIP from the BYOIP pool:

```hcl
resource "aws_eip" "byoip-ip" {
  vpc              = true
  public_ipv4_pool = "ipv4pool-ec2-012345"
}
```

## Argument Reference

The following arguments are supported:

* `vpc` - (Optional) Boolean if the EIP is in a VPC or not.
* `instance` - (Optional) EC2 instance ID.
* `network_interface` - (Optional) Network interface ID to associate with.
* `associate_with_private_ip` - (Optional) A user specified primary or secondary private IP address to
  associate with the Elastic IP address. If no private IP address is specified,
  the Elastic IP address is associated with the primary private IP address.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `public_ipv4_pool` - (Optional) EC2 IPv4 address pool identifier or `amazon`. This option is only available for VPC EIPs.

~> **NOTE:** You can specify either the `instance` ID or the `network_interface` ID,
but not both. Including both will **not** return an error from the AWS API, but will
have undefined behavior. See the relevant [AssociateAddress API Call][1] for
more information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Contains the EIP allocation ID.
* `private_ip` - Contains the private IP address (if in VPC).
* `associate_with_private_ip` - Contains the user specified private IP address
(if in VPC).
* `public_ip` - Contains the public IP address.
* `instance` - Contains the ID of the attached instance.
* `network_interface` - Contains the ID of the attached network interface.
* `public_ipv4_pool` - EC2 IPv4 address pool identifier (if in VPC).

## Timeouts
`aws_eip` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `read` - (Default `15 minutes`) How long to wait querying for information about EIPs.
- `update` - (Default `5 minutes`) How long to wait for an EIP to be updated.
- `delete` - (Default `3 minutes`) How long to wait for an EIP to be deleted.

## Import

EIPs in a VPC can be imported using their Allocation ID, e.g.

```
$ terraform import aws_eip.bar eipalloc-00a10e96
```

EIPs in EC2 Classic can be imported using their Public IP, e.g.

```
$ terraform import aws_eip.bar 52.0.0.0
```

[1]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AssociateAddress.html
