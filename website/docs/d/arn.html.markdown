---
layout: "aws"
page_title: "AWS: aws_arn"
sidebar_current: "docs-aws-datasource-arn"
description: |-
    Parses an ARN into its constituent parts.
---

# Data Source: aws_arn

Parses an Amazon Resource Name (ARN) into its constituent parts.

## Example Usage

```hcl
data "aws_arn" "db_instance" {
  arn = "arn:aws:rds:eu-west-1:123456789012:db:mysql-db"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Required) The ARN to parse.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `partition` - The partition that the resource is in.

* `service` - The [service namespace](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#genref-aws-service-namespaces) that identifies the AWS product.

* `region` - The region the resource resides in.
Note that the ARNs for some resources do not require a region, so this component might be omitted.

* `account` - The [ID](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html) of the AWS account that owns the resource, without the hyphens.

* `resource` - The content of this part of the ARN varies by service.
It often includes an indicator of the type of resource—for example, an IAM user or Amazon RDS database —followed by a slash (/) or a colon (:), followed by the resource name itself.
