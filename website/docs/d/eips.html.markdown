---
layout: "aws"
page_title: "AWS: aws_eips"
sidebar_current: "docs-aws-datasource-eips"
description: |-
   Use this data source to get ALLOCATION_IDs or PUBLIC_IPs of Amazon EIPs to be referenced elsewhere.
---

# Data Source: aws_eips

`aws_eips`Use this data source to get ALLOCATION_IDs or PUBLIC_IPS of Amazon EIPs to be referenced elsewhere.

## Example Usage

### Search By Tags (EC2-Classic or VPC)

```hcl
data "aws_eips" "by_tags" {
  tags = {
    Name = "exampleNameTagValue"
  }
}
```

## Argument Reference

* `tags` - (Required) A mapping of tags, each pair of which must exactly match a pair on the desired Elastic IPs

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - If VPC Elastic IP, the allocations identifiers. If EC2-Classic Elastic IPs, the public IPs addresses.
* `public_ips` - Public IP addresess of Elastic IPs.
