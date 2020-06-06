---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_ip_set"
description: |-
  Provides an AWS WAFv2 IP Set resource.
---

# Resource: aws_wafv2_ip_set

Provides a WAFv2 IP Set Resource

## Example Usage

```hcl
resource "aws_wafv2_ip_set" "example" {
  name               = "example"
  description        = "Example IP set"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name of the IP set.
* `description` - (Optional) A friendly description of the IP set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the Region US East (N. Virginia).
* `ip_address_version` - (Required) Specify IPV4 or IPV6. Valid values are `IPV4` or `IPV6`.
* `addresses` - (Required) Contains an array of strings that specify one or more IP addresses or blocks of IP addresses in Classless Inter-Domain Routing (CIDR) notation. AWS WAF supports all address ranges for IP versions IPv4 and IPv6.
* `tags` - (Optional) An array of key:value pairs to associate with the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for the set.
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster.

## Import

WAFv2 IP Sets can be imported using `ID/name/scope`

```
$ terraform import aws_wafv2_ip_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
