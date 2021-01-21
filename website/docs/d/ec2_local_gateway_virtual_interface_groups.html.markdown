---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_virtual_interface_groups"
description: |-
    Provides details about multiple EC2 Local Gateway Virtual Interface Groups
---

# Data Source: aws_ec2_local_gateway_virtual_interface_groups

Provides details about multiple EC2 Local Gateway Virtual Interface Groups, such as identifiers. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-networking-components.html#routing).

## Example Usage

```hcl
data "aws_ec2_local_gateway_virtual_interface_groups" "all" {}
```

## Argument Reference

The following arguments are optional:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGatewayVirtualInterfaceGroups.html) for supported filters. Detailed below.
* `tags` - (Optional) Key-value map of resource tags, each pair of which must exactly match a pair on the desired local gateway route table.

### filter Argument Reference

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `ids` - Set of EC2 Local Gateway Virtual Interface Group identifiers.
* `local_gateway_virtual_interface_ids` - Set of EC2 Local Gateway Virtual Interface identifiers.
