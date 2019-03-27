---
layout: "aws"
page_title: "AWS: aws_default_db_subnet_group"
sidebar_current: "docs-aws-resource-default-db-subnet-group"
description: |-
  Manage a default DB subnet group resource.
---

# aws_default_db_subnet_group

Provides a resource to manage a default DB subnet group in the current region.

The `aws_default_db_subnet_group` behaves differently from normal resources, in that 
Terraform does not _create_ this resource, but 
instead "adopts" it into management.

### Removing `aws_default_db_subnet_group` from your configuration

The `aws_default_db_subnet_group` resource allows you to manage a region's default DB subnet group,
but Terraform cannot destroy it. Removing this resource from your configuration will remove it from your statefile and management, but will not destroy the subnet group. You can resume managing the subnet group via the AWS Console.

## Example Usage

```hcl
resource "aws_default_db_subnet_group" "default" {
  subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
  tags {
    Name = "Default DB subnet group"
  }
}
```
## Argument Reference

The arguments of an `aws_default_db_subnet_group` differ from `aws_db_subnet_group` resources.
Namely, the `name` argument is computed.

The following arguments are still supported:

* `description` - (Optional) A description of the subnet group.
* `subnet_ids` – (Required) A list of VPC subnet IDs for the subnet group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The DB subnet group name.
* `arn` - The ARN of the DB subnet group.