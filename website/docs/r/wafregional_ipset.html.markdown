---
layout: "aws"
page_title: "AWS: wafregional_ipset"
sidebar_current: "docs-aws-resource-wafregional-ipset"
description: |-
  Provides a AWS WAF Regional IPSet resource for use with ALB.
---

# aws_wafregional_ipset

Provides a WAF Regional IPSet Resource for use with Application Load Balancer.

## Example Usage

```hcl
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }

  ip_set_descriptor {
    type  = "IPV4"
    value = "10.16.16.0/16"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the IPSet.
* `ip_set_descriptor` - (Optional) One or more pairs specifying the IP address type (IPV4 or IPV6) and the IP address range (in CIDR notation) from which web requests originate.

## Nested Blocks

### `ip_set_descriptor`

#### Arguments

* `type` - (Required) The string like IPV4 or IPV6.
* `value` - (Required) The CIDR notation.


## Remarks

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF IPSet.
* `arn` - The ARN of the WAF IPSet.
