---
subcategory: "Device Farm"
layout: "aws"
page_title: "AWS: aws_devicefarm_project"
description: |-
  Provides a Devicefarm project
---

# Resource: aws_devicefarm_project

Provides a resource to manage AWS Device Farm Projects.

For more information about Device Farm Projects, see the AWS Documentation on
[Device Farm Projects][aws-get-project].

~> **NOTE:** AWS currently has limited regional support for Device Farm (e.g. `us-west-2`). See [AWS Device Farm endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/devicefarm.html) for information on supported regions.

## Example Usage


```terraform
resource "aws_devicefarm_project" "awesome_devices" {
  name = "my-device-farm"
}
```

## Argument Reference

* `name` - (Required) The name of the project
* `default_job_timeout_minutes` - (Optional) Sets the execution timeout value (in minutes) for a project. All test runs in this project use the specified execution timeout value unless overridden when scheduling a run.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name of this project
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

[aws-get-project]: http://docs.aws.amazon.com/devicefarm/latest/APIReference/API_GetProject.html

## Import

DeviceFarm Projects can be imported by their arn:

```
$ terraform import aws_devicefarm_project.example arn:aws:devicefarm:us-west-2:123456789012:project:4fa784c7-ccb4-4dbf-ba4f-02198320daa1
```
