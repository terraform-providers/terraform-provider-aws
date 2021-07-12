---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_tag"
description: |-
  Manages an individual ECS resource tag
---

# Resource: aws_ecs_tag

Manages an individual ECS resource tag. This resource should only be used in cases where ECS resources are created outside Terraform (e.g. ECS Clusters implicitly created by Batch Compute Environments).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_ecs_cluster` and `aws_ecs_tag` to manage tags of the same ECS Cluster will cause a perpetual difference where the `aws_ecs_cluster` resource will try to remove the tag being added by the `aws_ecs_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_batch_compute_environment" "example" {
  compute_environment_name = "example"
  service_role             = aws_iam_role.example.arn
  type                     = "UNMANAGED"
}

resource "aws_ecs_tag" "example" {
  resource_arn = aws_batch_compute_environment.example.ecs_cluster_arn
  key          = "Name"
  value        = "Hello World"
}
```

## Argument Reference

The following arguments are supported:

* `resource_arn` - (Required) Amazon Resource Name (ARN) of the ECS resource to tag.
* `key` - (Required) Tag name.
* `value` - (Required) Tag value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ECS resource identifier and key, separated by a comma (`,`)

## Import

`aws_ecs_tag` can be imported by using the ECS resource identifier and key, separated by a comma (`,`), e.g.

```
$ terraform import aws_ecs_tag.example arn:aws:ecs:us-east-1:123456789012:cluster/example,Name
```
