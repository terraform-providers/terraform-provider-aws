---
subcategory: "File System (FSx)"
layout: "aws"
page_title: "AWS: aws_fsx_windows_file_system"
description: |-
  Manages a FSx Windows File System.
---

# Resource: aws_fsx_windows_file_system

Manages a FSx Windows File System. See the [FSx Windows Guide](https://docs.aws.amazon.com/fsx/latest/WindowsGuide/what-is.html) for more information.

~> **NOTE:** Either the `active_directory_id` argument or `self_managed_active_directory` configuration block must be specified.

## Example Usage

### Using AWS Directory Service

Additional information for using AWS Directory Service with Windows File Systems can be found in the [FSx Windows Guide](https://docs.aws.amazon.com/fsx/latest/WindowsGuide/fsx-aws-managed-ad.html).

```hcl
resource "aws_fsx_windows_file_system" "example" {
  active_directory_id = aws_directory_service_directory.example.id
  kms_key_id          = aws_kms_key.example.arn
  storage_capacity    = 300
  subnet_ids          = [aws_subnet.example.id]
  throughput_capacity = 1024
}
```

### Using a Self-Managed Microsoft Active Directory

Additional information for using AWS Directory Service with Windows File Systems can be found in the [FSx Windows Guide](https://docs.aws.amazon.com/fsx/latest/WindowsGuide/self-managed-AD.html).

```hcl
resource "aws_fsx_windows_file_system" "example" {
  kms_key_id          = aws_kms_key.example.arn
  storage_capacity    = 300
  subnet_ids          = [aws_subnet.example.id]
  throughput_capacity = 1024

  self_managed_active_directory {
    dns_ips     = ["10.0.0.111", "10.0.0.222"]
    domain_name = "corp.example.com"
    password    = "avoid-plaintext-passwords"
    username    = "Admin"
  }
}
```

## Argument Reference

The following arguments are supported:

* `storage_capacity` - (Required) Storage capacity (GiB) of the file system. Minimum of 32 and maximum of 65536. If the storage type is set to `HDD` the minimum value is 2000.
* `subnet_ids` - (Required) A list of IDs for the subnets that the file system will be accessible from. To specify more than a single subnet set `deployment_type` to `MULTI_AZ_1`.
* `throughput_capacity` - (Required) Throughput (megabytes per second) of the file system in power of 2 increments. Minimum of `8` and maximum of `2048`.
* `active_directory_id` - (Optional) The ID for an existing Microsoft Active Directory instance that the file system should join when it's created. Cannot be specified with `self_managed_active_directory`.
* `automatic_backup_retention_days` - (Optional) The number of days to retain automatic backups. Minimum of `0` and maximum of `90`. Defaults to `7`. Set to `0` to disable.
* `copy_tags_to_backups` - (Optional) A boolean flag indicating whether tags on the file system should be copied to backups. Defaults to `false`.
* `daily_automatic_backup_start_time` - (Optional) The preferred time (in `HH:MM` format) to take daily automatic backups, in the UTC time zone.
* `kms_key_id` - (Optional) ARN for the KMS Key to encrypt the file system at rest. Defaults to an AWS managed KMS Key.
* `security_group_ids` - (Optional) A list of IDs for the security groups that apply to the specified network interfaces created for file system access. These security groups will apply to all network interfaces.
* `self_managed_active_directory` - (Optional) Configuration block that Amazon FSx uses to join the Windows File Server instance to your self-managed (including on-premises) Microsoft Active Directory (AD) directory. Cannot be specified with `active_directory_id`. Detailed below.
* `skip_final_backup` - (Optional) When enabled, will skip the default final backup taken when the file system is deleted. This configuration must be applied separately before attempting to delete the resource to have the desired behavior. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the file system.
* `weekly_maintenance_start_time` - (Optional) The preferred start time (in `d:HH:MM` format) to perform weekly maintenance, in the UTC time zone.
* `deployment_type` - (Optional) Specifies the file system deployment type, valid values are `MULTI_AZ_1` and `SINGLE_AZ_1`. Default value is `SINGLE_AZ_1`.
* `preferred_subnet_id` - (Optional) Specifies the subnet in which you want the preferred file server to be located. Required for when deployment type is `MULTI_AZ_1`.
* `storage_type` - (Optional) Specifies the storage type, Valid values are `SSD` and `HDD`. `HDD` is supported on `SINGLE_AZ_1` and `MULTI_AZ_1` Windows file system deployment types. Default value is `SSD`.

### self_managed_active_directory

The following arguments are supported for `self_managed_active_directory` configuration block:

* `dns_ips` - (Required) A list of up to two IP addresses of DNS servers or domain controllers in the self-managed AD directory. The IP addresses need to be either in the same VPC CIDR range as the file system or in the private IP version 4 (IPv4) address ranges as specified in [RFC 1918](https://tools.ietf.org/html/rfc1918).
* `domain_name` - (Required) The fully qualified domain name of the self-managed AD directory. For example, `corp.example.com`.
* `password` - (Required) The password for the service account on your self-managed AD domain that Amazon FSx will use to join to your AD domain.
* `username` - (Required) The user name for the service account on your self-managed AD domain that Amazon FSx will use to join to your AD domain.
* `file_system_administrators_group` - (Optional) The name of the domain group whose members are granted administrative privileges for the file system. Administrative privileges include taking ownership of files and folders, and setting audit controls (audit ACLs) on files and folders. The group that you specify must already exist in your domain. Defaults to `Domain Admins`.
* `organizational_unit_distinguished_name` - (Optional) The fully qualified distinguished name of the organizational unit within your self-managed AD directory that the Windows File Server instance will join. For example, `OU=FSx,DC=yourdomain,DC=corp,DC=com`. Only accepts OU as the direct parent of the file system. If none is provided, the FSx file system is created in the default location of your self-managed AD directory. To learn more, see [RFC 2253](https://tools.ietf.org/html/rfc2253).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `dns_name` - DNS name for the file system, e.g. `fs-12345678.corp.example.com` (domain name matching the Active Directory domain name)
* `id` - Identifier of the file system, e.g. `fs-12345678`
* `network_interface_ids` - Set of Elastic Network Interface identifiers from which the file system is accessible.
* `owner_id` - AWS account identifier that created the file system.
* `vpc_id` - Identifier of the Virtual Private Cloud for the file system.
* `preferred_file_server_ip` - The IP address of the primary, or preferred, file server.
* `remote_administration_endpoint` - For `MULTI_AZ_1` deployment types, use this endpoint when performing administrative tasks on the file system using Amazon FSx Remote PowerShell. For `SINGLE_AZ_1` deployment types, this is the DNS name of the file system.

## Timeouts

`aws_fsx_windows_file_system` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

* `create` - (Default `30m`) How long to wait for the file system to be created.
* `delete` - (Default `30m`) How long to wait for the file system to be deleted.

## Import

FSx File Systems can be imported using the `id`, e.g.

```
$ terraform import aws_fsx_windows_file_system.example fs-543ab12b1ca672f33
```

Certain resource arguments, like `security_group_ids` and the `self_managed_active_directory` configuation block `password`, do not have a FSx API method for reading the information after creation. If these arguments are set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](/docs/configuration/resources.html#ignore_changes) to hide the difference, e.g.

```hcl
resource "aws_fsx_windows_file_system" "example" {
  # ... other configuration ...

  security_group_ids = [aws_security_group.example.id]

  # There is no FSx API for reading security_group_ids
  lifecycle {
    ignore_changes = [security_group_ids]
  }
}
```
