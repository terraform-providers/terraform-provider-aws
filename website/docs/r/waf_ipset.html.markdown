---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_waf_ipset"
description: |-
  Provides a AWS WAF IPSet resource.
---

# Resource: aws_waf_ipset

Provides a WAF IPSet Resource

## Example Usage

```hcl
resource "aws_waf_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }

  ip_set_descriptors {
    type  = "IPV4"
    value = "10.16.16.0/16"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the IPSet.
* `ip_set_descriptors` - (Optional) One or more pairs specifying the IP address type (IPV4 or IPV6) and the IP address range (in CIDR format) from which web requests originate.

## Nested Blocks

### `ip_set_descriptors`

#### Arguments

* `type` - (Required) Type of the IP address - `IPV4` or `IPV6`.
* `value` - (Required) An IPv4 or IPv6 address specified via CIDR notation. e.g. `192.0.2.44/32` or `1111:0000:0000:0000:0000:0000:0000:0000/64`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF IPSet.
* `arn` - The ARN of the WAF IPSet.

## Import

WAF IPSets can be imported using their ID, e.g.

```
$ terraform import aws_waf_ipset.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
