---
subcategory: "Application Autoscaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_policy"
description: |-
  Provides an Application AutoScaling Policy data source.
---

# aws_appautoscaling_policy

Provides information about an Application AutoScaling Policy.

## Example Usage

```hcl
variable "name" {
  type = "string"
}

variable "service_namespace" {
  type = "string"
}

data "aws_appautoscaling_policy" "test" {
  name              = var.name
  service_namespace = var.service_namespace
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the policy.
* `service_namespace` - (Required) The AWS service namespace of the scalable target. Documentation can be found in the `ServiceNamespace` parameter at: [AWS Application Auto Scaling API Reference](http://docs.aws.amazon.com/ApplicationAutoScaling/latest/APIReference/API_RegisterScalableTarget.html#API_RegisterScalableTarget_RequestParameters)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) identifying your policy.
* `policy_type` - Type of scaling policy.
* `resource_id` - The resource type and unique identifier string for the resource associated with the scaling policy.
* `scalable_dimension` - The scalable dimension of the scalable target.