---
subcategory: "Global Accelerator"
layout: "aws"
page_title: "AWS: aws_globalaccelerator_accelerator"
description: |-
  Provides a Global Accelerator accelerator.
---

# Resource: aws_globalaccelerator_accelerator

Creates a Global Accelerator accelerator.

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
* `tags` - (Optional) A map of tags to assign to the resource.

**attributes** supports the following attributes:

* `flow_logs_enabled` - (Optional) Indicates whether flow logs are enabled.
* `flow_logs_s3_bucket` - (Optional) The name of the Amazon S3 bucket for the flow logs.
* `flow_logs_s3_prefix` - (Optional) The prefix for the location in the Amazon S3 bucket for the flow logs.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the accelerator.
* `dns_name` - The DNS name of the accelerator. For example, `a5d53ff5ee6bca4ce.awsglobalaccelerator.com`.
* `hosted_zone_id` --  The Global Accelerator Route 53 zone ID that can be used to
  route an [Alias Resource Record Set][1] to the Global Accelerator. This attribute
  is simply an alias for the zone ID `Z2BJ6XQ5FK7U4H`.
* `ip_sets` - IP address set associated with the accelerator.

**ip_sets** exports the following attributes:

* `ip_addresses` - A list of IP addresses in the IP address set.
* `ip_family` - The types of IP addresses included in this IP set.

[1]: https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html

## Timeouts

`aws_globalaccelerator_accelerator` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `30 minutes`) How long to wait for the Global Accelerator Accelerator to be created.
* `update` - (Default `30 minutes`) How long to wait for the Global Accelerator Accelerator to be updated.

## Import

Global Accelerator accelerators can be imported using the `id`, e.g.

```
$ terraform import aws_globalaccelerator_accelerator.example arn:aws:globalaccelerator::111111111111:accelerator/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```
