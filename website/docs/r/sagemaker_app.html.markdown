---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_app"
description: |-
  Provides a Sagemaker App resource.
---

# Resource: aws_sagemaker_app

Provides a Sagemaker App resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_app" "example" {
  domain_id         = aws_sagemaker_domain.example.id
  user_profile_name = aws_sagemaker_user_profile.example.user_profile_name
  app_name          = "example"
  app_type          = "JupyterServer"
}
```

## Argument Reference

The following arguments are supported:

* `app_name` - (Required) The name of the app.
* `app_type` - (Required) The type of app. Valid values are `JupyterServer`, `KernelGateway` and `TensorBoard`.
* `domain_id` - (Required) The domain ID.
* `user_profile_name` - (Required) The user profile name.
* `resource_spec` - (Optional) The instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance.See [Resource Spec](#resource-spec) below.
* `tags` - (Optional) A map of tags to assign to the resource.

### Resource Spec

* `instance_type` - (Optional) The instance type that the image version runs on. For valid values see [Sagemaker Instance Types](https://docs.aws.amazon.com/sagemaker/latest/dg/notebooks-available-instance-types.html).
* `sagemaker_image_arn` - (Optional) The ARN of the SageMaker image that the image version belongs to.
* `sagemaker_image_version_arn` - (Optional) The ARN of the image version created on the instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the app.
* `arn` - The Amazon Resource Name (ARN) of the app.

## Import

Sagemaker Code Apps can be imported using the `id`, e.g.

```
$ terraform import aws_sagemaker_app.example arn:aws:sagemaker:us-west-2:012345678912:app/domain-id/user-profile-name/app-type/app-name
```
