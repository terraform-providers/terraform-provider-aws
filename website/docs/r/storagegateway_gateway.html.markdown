---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_gateway"
description: |-
  Manages an AWS Storage Gateway file, tape, or volume gateway in the provider region
---

# Resource: aws_storagegateway_gateway

Manages an AWS Storage Gateway file, tape, or volume gateway in the provider region.

~> NOTE: The Storage Gateway API requires the gateway to be connected to properly return information after activation. If you are receiving `The specified gateway is not connected` errors during resource creation (gateway activation), ensure your gateway instance meets the [Storage Gateway requirements](https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html).

## Example Usage

### File Gateway

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
```

### Tape Gateway

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address  = "1.2.3.4"
  gateway_name        = "example"
  gateway_timezone    = "GMT"
  gateway_type        = "VTL"
  medium_changer_type = "AWS-Gateway-VTL"
  tape_drive_type     = "IBM-ULT3580-TD5"
}
```

### Volume Gateway (Cached)

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"
}
```

### Volume Gateway (Stored)

```hcl
resource "aws_storagegateway_gateway" "example" {
  gateway_ip_address = "1.2.3.4"
  gateway_name       = "example"
  gateway_timezone   = "GMT"
  gateway_type       = "STORED"
}
```

## Argument Reference

~> **NOTE:** One of `activation_key` or `gateway_ip_address` must be provided for resource creation (gateway activation). Neither is required for resource import. If using `gateway_ip_address`, Terraform must be able to make an HTTP (port 80) GET request to the specified IP address from where it is running.

The following arguments are supported:

* `gateway_name` - (Required) Name of the gateway.
* `gateway_timezone` - (Required) Time zone for the gateway. The time zone is of the format "GMT", "GMT-hr:mm", or "GMT+hr:mm". For example, `GMT-4:00` indicates the time is 4 hours behind GMT. The time zone is used, for example, for scheduling snapshots and your gateway's maintenance schedule.
* `activation_key` - (Optional) Gateway activation key during resource creation. Conflicts with `gateway_ip_address`. Additional information is available in the [Storage Gateway User Guide](https://docs.aws.amazon.com/storagegateway/latest/userguide/get-activation-key.html).
* `average_download_rate_limit_in_bits_per_sec` - (Optional) The average download bandwidth rate limit in bits per second. This is supported for the `CACHED`, `STORED`, and `VTL` gateway types.
* `average_upload_rate_limit_in_bits_per_sec` - (Optional) The average upload bandwidth rate limit in bits per second. This is supported for the `CACHED`, `STORED`, and `VTL` gateway types.
* `gateway_ip_address` - (Optional) Gateway IP address to retrieve activation key during resource creation. Conflicts with `activation_key`. Gateway must be accessible on port 80 from where Terraform is running. Additional information is available in the [Storage Gateway User Guide](https://docs.aws.amazon.com/storagegateway/latest/userguide/get-activation-key.html).
* `gateway_type` - (Optional) Type of the gateway. The default value is `STORED`. Valid values: `CACHED`, `FILE_S3`, `STORED`, `VTL`.
* `gateway_vpc_endpoint` - (Optional) VPC endpoint address to be used when activating your gateway. This should be used when your instance is in a private subnet. Requires HTTP access from client computer running terraform. More info on what ports are required by your VPC Endpoint Security group in [Activating a Gateway in a Virtual Private Cloud](https://docs.aws.amazon.com/storagegateway/latest/userguide/gateway-private-link.html).
* `cloudwatch_log_group_arn` - (Optional) The Amazon Resource Name (ARN) of the Amazon CloudWatch log group to use to monitor and log events in the gateway.
* `medium_changer_type` - (Optional) Type of medium changer to use for tape gateway. Terraform cannot detect drift of this argument. Valid values: `STK-L700`, `AWS-Gateway-VTL`, `IBM-03584L32-0402`.
* `smb_active_directory_settings` - (Optional) Nested argument with Active Directory domain join information for Server Message Block (SMB) file shares. Only valid for `FILE_S3` gateway type. Must be set before creating `ActiveDirectory` authentication SMB file shares. More details below.
* `smb_guest_password` - (Optional) Guest password for Server Message Block (SMB) file shares. Only valid for `FILE_S3` gateway type. Must be set before creating `GuestAccess` authentication SMB file shares. Terraform can only detect drift of the existence of a guest password, not its actual value from the gateway. Terraform can however update the password with changing the argument.
* `smb_security_strategy` - (Optional) Specifies the type of security strategy. Valid values are: `ClientSpecified`, `MandatorySigning`, and `MandatoryEncryption`. See [Setting a Security Level for Your Gateway](https://docs.aws.amazon.com/storagegateway/latest/userguide/managing-gateway-file.html#security-strategy) for more information.
* `tape_drive_type` - (Optional) Type of tape drive to use for tape gateway. Terraform cannot detect drift of this argument. Valid values: `IBM-ULT3580-TD5`.
* `tags` - (Optional) Key-value map of resource tags

### smb_active_directory_settings

Information to join the gateway to an Active Directory domain for Server Message Block (SMB) file shares.

~> **NOTE** It is not possible to unconfigure this setting without recreating the gateway. Also, Terraform can only detect drift of the `domain_name` argument from the gateway.

* `domain_name` - (Required) The name of the domain that you want the gateway to join.
* `password` - (Required) The password of the user who has permission to add the gateway to the Active Directory domain.
* `username` - (Required) The user name of user who has permission to add the gateway to the Active Directory domain.
* `timeout_in_seconds` - (Optional) Specifies the time in seconds, in which the JoinDomain operation must complete. The default is `20` seconds.
* `organizational_unit` - (Optional) The organizational unit (OU) is a container in an Active Directory that can hold users, groups,
 computers, and other OUs and this parameter specifies the OU that the gateway will join within the AD domain.
* `domain_controllers` - (Optional) List of IPv4 addresses, NetBIOS names, or host names of your domain server.
 If you need to specify the port number include it after the colon (“:”). For example, `mydc.mydomain.com:389`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the gateway.
* `arn` - Amazon Resource Name (ARN) of the gateway.
* `gateway_id` - Identifier of the gateway.
* `ec2_instance_id` - The ID of the Amazon EC2 instance that was used to launch the gateway.
* `endpoint_type` - The type of endpoint for your gateway.
* `host_environment` - The type of hypervisor environment used by the host.
* `gateway_network_interface` - An array that contains descriptions of the gateway network interfaces. See [Gateway Network Interface](#gateway-network-interface).

### Gateway Network Interface

* `ipv4_address` - The Internet Protocol version 4 (IPv4) address of the interface.

## Timeouts

`aws_storagegateway_gateway` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for gateway activation and connection to Storage Gateway.

## Import

`aws_storagegateway_gateway` can be imported by using the gateway Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_gateway.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678
```

Certain resource arguments, like `gateway_ip_address` do not have a Storage Gateway API method for reading the information after creation, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.


```hcl
resource "aws_storagegateway_gateway" "example" {
  # ... other configuration ...

  gateway_ip_address = aws_instance.sgw.private_ip
  # There is no Storage Gateway API for reading gateway_ip_address
  lifecycle {
    ignore_changes = ["gateway_ip_address"]
  }
}
```
