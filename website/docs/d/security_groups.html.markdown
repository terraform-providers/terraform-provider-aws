---
layout: "aws"
page_title: "AWS: aws_security_groups"
sidebar_current: "docs-aws-datasource-security-groups"
description: |-
  Get information about a set of Security Groups.
---

# Data Source: aws_security_groups

Use this data source to get IDs and VPC membership of Security Groups that are created
outside of Terraform.

## Example Usage
```hcl
data "aws_security_groups" "test" {
  tags = {
    Application = "k8s"
    Environment = "dev"
  }
}
```

```hcl
data "aws_security_groups" "test" {
  filter {
    name   = "group-name"
    values = ["*nodes*"]
  }

  filter {
    name   = "vpc-id"
    values = ["${var.vpc_id}"]
  }
}
```

## Argument Reference

* `tags` - (Optional) A mapping of tags, each pair of which must exactly match for
desired security groups.

* `filter` - (Optional) One or more name/value pairs to use as filters. There are
several valid keys, for a full reference, check out
[describe-security-groups in the AWS CLI reference][1].

## Attributes Reference

* `ids` - IDs of the matches security groups.
* `vpc_ids` - The VPC IDs of the matched security groups. The data source's tag or filter *will span VPCs*
unless the `vpc-id` filter is also used.

[1]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-security-groups.html
