---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_subnet_group"
description: |-
  Provides an ElastiCache Subnet Group resource.
---

# Resource: aws_elasticache_subnet_group

Provides an ElastiCache Subnet Group resource.

~> **NOTE:** ElastiCache Subnet Groups are only for use when working with an
ElastiCache cluster **inside** of a VPC. If you are on EC2 Classic, see the
[ElastiCache Security Group resource](elasticache_security_group.html).

## Example Usage

```terraform
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-test"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-test"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name       = "tf-test-cache-subnet"
  subnet_ids = [aws_subnet.foo.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) Name for the cache subnet group. Elasticache converts this name to lowercase.
* `description` – (Optional) Description for the cache subnet group. Defaults to "Managed by Terraform".
* `subnet_ids` – (Required) List of VPC Subnet IDs for the cache subnet group
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `description` - The Description of the ElastiCache Subnet Group.
* `name` - The Name of the ElastiCache Subnet Group.
* `subnet_ids` - The Subnet IDs of the ElastiCache Subnet Group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).


## Import

ElastiCache Subnet Groups can be imported using the `name`, e.g.

```
$ terraform import aws_elasticache_subnet_group.bar tf-test-cache-subnet
```
