---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_auto_scaling_configuration_version"
description: |-
  Manages an App Runner AutoScaling Configuration Version.
---

# Resource: aws_apprunner_auto_scaling_configuration_version

Manages an App Runner AutoScaling Configuration Version.

## Example Usage

```terraform
resource "aws_apprunner_auto_scaling_configuration_version" "example" {
  auto_scaling_configuration_name = "example"

  max_concurrency = 50
  max_size        = 10
  min_size        = 2

  tags = {
    Name = "example-apprunner-autoscaling"
  }
}
```

## Argument Reference

The following arguments supported:

* `auto_scaling_configuration_name` - (Required, Forces new resource) Name of the auto scaling configuration.
* `max_concurrency` - (Optional, Forces new resource) The maximal number of concurrent requests that you want an instance to process. When the number of concurrent requests goes over this limit, App Runner scales up your service.
* `max_size` - (Optional, Forces new resource) The maximal number of instances that App Runner provisions for your service.
* `min_size` - (Optional, Forces new resource) The minimal number of instances that App Runner provisions for your service.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of this auto scaling configuration version.
* `auto_scaling_configuration_revision` - The revision of this auto scaling configuration.
* `latest` - Whether the auto scaling configuration has the highest `auto_scaling_configuration_revision` among all configurations that share the same `auto_scaling_configuration_name`.
* `status` - The current state of the auto scaling configuration. An INACTIVE configuration revision has been deleted and can't be used. It is permanently removed some time after deletion.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

App Runner AutoScaling Configuration Versions can be imported by using the `arn`, e.g.

```
$ terraform import aws_apprunner_auto_scaling_configuration_version.example "arn:aws:apprunner:us-east-1:1234567890:autoscalingconfiguration/example/1/69bdfe0115224b0db49398b7beb68e0f
```
