---
layout: "aws"
page_title: "AWS: aws_globalaccelerator_accelerator"
sidebar_current: "docs-aws-resource-globalaccelerator-accelerator"
description: |-
  Provides a Global Accelerator accelerator.
---

# aws_globalaccelerator_accelerator

Provides a Global Accelerator accelerator.

## Example Usage

```hcl
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  enabled         = true

  attributes {
    flow_logs_enabled   = true
    flow_logs_s3_bucket = "example-bucket"
    flow_logs_s3_prefix = "flow-logs/"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the accelerator.
* `ip_address_type` - (Optional) The value for the address type must be `IPV4`.
* `enabled` - (Optional) Indicates whether the accelerator is enabled. The value is true or false. The default value is true.
* `attributes` - (Optional) The attributes of the accelerator. Fields documented below.

**attributes** supports the following attributes:

* `flow_logs_enabled` - (Optional) Indicates whether flow logs are enabled.
* `flow_logs_s3_bucket` - (Optional) The name of the Amazon S3 bucket for the flow logs.
* `flow_logs_s3_prefix` - (Optional) The prefix for the location in the Amazon S3 bucket for the flow logs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the accelerator.
* `ip_sets` - IP address set associated with the accelerator.

**ip_sets** exports the following attributes:

* `ip_addresses` - The array of IP addresses in the IP address set.
* `ip_family` - The types of IP addresses included in this IP set.

## Import

Global Accelerator accelerators can be imported using the `id`, e.g.

```
$ terraform import aws_globalaccelerator_accelerator.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
