---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_directory"
description: |-
  AWS Directory Service Directory
---

# Data Source: aws_directory_service_directory

Get attributes of AWS Directory Service directory (SimpleAD, Managed AD, AD Connector). It's especially useful to refer AWS Managed AD or on-premise AD in AD Connector configuration.

## Example Usage

```hcl
data "aws_directory_service_directory" "example" {
  directory_id = aws_directory_service_directory.main.id
}
```

## Argument Reference

* `directory_id` - (Required) The ID of the directory.

## Attributes Reference

* `type` - The directory type (`SimpleAD`, `ADConnector` or `MicrosoftAD`).
* `edition` - (for `MicrosoftAD`) The Microsoft AD edition (`Standard` or `Enterprise`).
* `name` - The fully qualified name for the directory/connector.
* `password` - The password for the directory administrator or connector user.
* `size` - (for `SimpleAD` and `ADConnector`) The size of the directory/connector (`Small` or `Large`).
* `alias` - The alias for the directory/connector, such as `d-991708b282.awsapps.com`.
* `description` - A textual description for the directory/connector.
* `short_name` - The short name of the directory/connector, such as `CORP`.
* `enable_sso` - The directory/connector single-sign on status.
* `access_url` - The access URL for the directory/connector, such as http://alias.awsapps.com.
* `dns_ip_addresses` - A list of IP addresses of the DNS servers for the directory/connector.
* `security_group_id` - The ID of the security group created by the directory/connector.
* `tags` – A map of tags assigned to the directory/connector.

 `vpc_settings` (for `SimpleAD` and `MicrosoftAD`) is also exported with the following attributes:

* `subnet_ids` - The identifiers of the subnets for the directory servers (2 subnets in 2 different AZs).
* `vpc_id` - The ID of the VPC that the directory is in.

`connect_settings` (for `ADConnector`) is also exported with the following attributes:

* `connect_ips` - The IP addresses of the AD Connector servers.
* `customer_username` - The username corresponding to the password provided.
* `customer_dns_ips` - The DNS IP addresses of the domain to connect to.
* `subnet_ids` - The identifiers of the subnets for the connector servers (2 subnets in 2 different AZs).
* `vpc_id` - The ID of the VPC that the connector is in.
