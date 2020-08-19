---
subcategory: "MediaStore"
layout: "aws"
page_title: "AWS: aws_media_store_container_metric_policy"
description: |-
  Provides a MediaStore Container Metric Policy.
---

# Resource: aws_media_store_container_metric_policy

Provides a MediaStore Container Metric Policy.

## Example Usage

```hcl
resource "aws_media_store_container" "example" {
  name = "example"
}

resource "aws_media_store_container_metric_policy" "example" {
  container_name = aws_media_store_container.example.name

  metric_policy {
    container_level_metrics = "DISABLED"

    metric_policy_rule {
      object_group      = "baseball/saturday"
      object_group_name = "baseballGroup"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `container_name` - (Required) The name of the container.
* `metric_policy` - (Required) Configure Metric Policy properties. See below for details.

The `metric_policy` block supports the following:

* `container_level_metrics` - (Required) A setting to enable or disable metrics at the container level.
* `metric_policy_rule` - (Optional) The rules that enable metrics at the object level. By default, you can include up to five rules. You can also request a quota increase (https://console.aws.amazon.com/servicequotas/home?region=us-east-1#!/services/mediastore/quotas) to allow up to 300 rules per policy. See below for details.

Each `metric_policy_rule` supports the following:

* `object_group` - (Required) A path or file name that defines which objects to include in the group. Wildcards (*) are acceptable.
* `object_group_name` - (Required) A name that allows you to refer to the object group.


## Import

MediaStore Container Metric Policy can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_container_metric_policy.example example
```
