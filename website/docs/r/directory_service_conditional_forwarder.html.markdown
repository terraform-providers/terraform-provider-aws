---
layout: "aws"
page_title: "AWS: aws_directory_service_conditional_forwarder"
sidebar_current: "docs-aws-resource-directory-service-conditional-forwarder"
description: |-
  Provides a conditional forwarder for managed Microsoft AD in AWS Directory Service.
---

# aws_directory_service_conditional_forwarder

Provides a conditional forwarder for managed Microsoft AD in AWS Directory Service.

## Example Usage

```hcl
resource "aws_directory_service_conditional_forwarder" "example" {
  directory_id       = "${aws_directory_service_directory.ad.id}"
  remote_domain_name = "example.com"

  dns_ips = [
    "8.8.8.8",
    "8.8.4.4",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The id of directory.
* `dns_ips` - (Required) A list of forwarder IP addresses.
* `remote_domain_name` - (Required) The fully qualified domain name of the remote domain for which forwarders will be used.

## Import

Conditional forwarders can be imported using the directory id and remote_domain_name, e.g.

```
$ terraform import aws_directory_service_conditional_forwarder.example d-1234567890:example.com
```
