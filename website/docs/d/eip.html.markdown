---
layout: "aws"
page_title: "AWS: aws_eip"
sidebar_current: "docs-aws-datasource-eip"
description: |-
    Provides details about a specific Elastic IP
---

# Data Source: aws_eip

`aws_eip` provides details about a specific Elastic IP.

## Example Usage

### Search By Allocation ID (VPC only)

```hcl
data "aws_eip" "by_allocation_id" {
  id = "eipalloc-12345678"
}
```

### Search By Filters (EC2-Classic or VPC)

```hcl
data "aws_eip" "by_filter" {
  filter {
    name   = "tag:Name"
    values = ["exampleNameTagValue"]
  }
}
```

### Search By Public IP (EC2-Classic or VPC)

```hcl
data "aws_eip" "by_public_ip" {
  public_ip = "1.2.3.4"
}
```

### Search By Tags (EC2-Classic or VPC)

```hcl
data "aws_eip" "by_tags" {
  tags = {
    Name = "exampleNameTagValue"
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
Elastic IPs in the current region. The given filters must match exactly one
Elastic IP whose data will be exported as attributes.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are several valid keys, for a full reference, check out the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAddresses.html).
* `id` - (Optional) The allocation id of the specific VPC EIP to retrieve. If a classic EIP is required, do NOT set `id`, only set `public_ip`
* `public_ip` - (Optional) The public IP of the specific EIP to retrieve.
* `tags` - (Optional) A mapping of tags, each pair of which must exactly match a pair on the desired Elastic IP

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `association_id` - The ID representing the association of the address with an instance in a VPC.
* `domain` - Indicates whether the address is for use in EC2-Classic (standard) or in a VPC (vpc).
* `id` - If VPC Elastic IP, the allocation identifier. If EC2-Classic Elastic IP, the public IP address.
* `instance_id` - The ID of the instance that the address is associated with (if any).
* `network_interface_id` - The ID of the network interface.
* `network_interface_owner_id` - The ID of the AWS account that owns the network interface.
* `private_ip` - The private IP address associated with the Elastic IP address.
* `public_ip` - Public IP address of Elastic IP.
* `public_ipv4_pool` - The ID of an address pool.
* `tags` - Key-value map of tags associated with Elastic IP.
