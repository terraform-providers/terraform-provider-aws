---
subcategory: "Autoscaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_group_tag"
description: |-
  Manages an individual Autoscaling Group tag
---

# Resource: aws_autoscaling_group_tag

Manages an individual Autoscaling Group (ASG) tag. This resource should only be used in cases where ASGs are created outside Terraform (e.g. ASGs implicitly created by EKS Node Groups).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_autoscaling_group` and `aws_autoscaling_group_tag` to manage tags of the same ASG will cause a perpetual difference where the `aws_autoscaling_group` resource will try to remove the tag being added by the `aws_autoscaling_group_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```terraform
resource "aws_eks_node_group" "example" {
  cluster_name = "example"
  node_group_name = "example"

  # ... other configuration ...
}

resource "aws_autoscaling_group_tag" "example" {
  for_each = toset(
    [for asg in flatten(
      [for resources in aws_eks_node_group.example.resources: resources.autoscaling_groups]
    ): asg.name]
  )

  asg_name = each.value

  tag {
    key   = "k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/capacityType"
    value = "SPOT"

    propagate_at_launch = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `asg_name` - (Required) The name of the Autoscaling Group to apply the tag to.
* `tag` - (Required) The tag to create. The `tag` block is documented below.

The `tag` block supports the following arguments:

* `key` - (Required) Tag name.
* `value` - (Required) Tag value.
* `propagate_at_launch` - (Required) Whether to propagate the tags to instances launched by the ASG.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ASG name and key, separated by a comma (`,`)

## Import

`aws_autoscaling_group_tag` can be imported by using the ASG name and key, separated by a comma (`,`), e.g.

```
$ terraform import aws_autoscaling_group.example asg-,k8s.io/cluster-autoscaler/node-template/label/eks.amazonaws.com/capacityType
```
