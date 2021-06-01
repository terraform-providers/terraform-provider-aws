---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_user_profile"
description: |-
  Provides a Sagemaker User Profile resource.
---

# Resource: aws_sagemaker_user_profile

Provides a Sagemaker User Profile resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_user_profile" "example" {
  domain_id         = aws_sagemaker_domain.test.id
  user_profile_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `user_profile_name` - (Required) The name for the User Profile.
* `domain_id` - (Required) The ID of the associated Domain.
* `single_sign_on_user_identifier` - (Optional) A specifier for the type of value specified in `single_sign_on_user_value`. Currently, the only supported value is `UserName`. If the Domain's AuthMode is SSO, this field is required. If the Domain's AuthMode is not SSO, this field cannot be specified.
* `single_sign_on_user_value` - (Required) The username of the associated AWS Single Sign-On User for this User Profile. If the Domain's AuthMode is SSO, this field is required, and must match a valid username of a user in your directory. If the Domain's AuthMode is not SSO, this field cannot be specified.
* `user_settings` - (Required) The user settings. See [User Settings](#user-settings) below.
* `tags` - (Optional) A map of tags to assign to the resource.

### User Settings

* `execution_role` - (Required) The execution role ARN for the user.
* `security_groups` - (Optional) The security groups.
* `sharing_settings` - (Optional) The sharing settings. See [Sharing Settings](#sharing-settings) below.
* `tensor_board_app_settings` - (Optional) The TensorBoard app settings. See [TensorBoard App Settings](#tensorboard-app-settings) below.
* `jupyter_server_app_settings` - (Optional) The Jupyter server's app settings. See [Jupyter Server App Settings](#jupyter-server-app-settings) below.
* `kernel_gateway_app_settings` - (Optional) The kernel gateway app settings. See [Kernel Gateway App Settings](#kernal-gateway-app-settings) below.

#### Sharing Settings

* `notebook_output_option` - (Optional) Whether to include the notebook cell output when sharing the notebook. The default is `Disabled`. Valid values are `Allowed` and `Disabled`.
* `s3_kms_key_id` - (Optional) When `notebook_output_option` is Allowed, the AWS Key Management Service (KMS) encryption key ID used to encrypt the notebook cell output in the Amazon S3 bucket.
* `s3_output_path` - (Optional) When `notebook_output_option` is Allowed, the Amazon S3 bucket used to save the notebook cell output.

#### TensorBoard App Settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default-resource-spec) below.

#### Kernel Gateway App Settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default-resource-spec) below.
* `custom_image` - (Optional) A list of custom SageMaker images that are configured to run as a KernelGateway app. see [Custom Image](#custom-image) below.

#### Jupyter Server App Settings

* `default_resource_spec` - (Optional) The default instance type and the Amazon Resource Name (ARN) of the SageMaker image created on the instance. see [Default Resource Spec](#default-resource-spec) below.

##### Default Resource Spec

* `instance_type` - (Optional) The instance type.
* `sagemaker_image_arn` - (Optional) The Amazon Resource Name (ARN) of the SageMaker image created on the instance.

##### Custom Image

* `app_image_config_name` - (Required) The name of the App Image Config.
* `image_name` - (Required) The name of the Custom Image.
* `image_version_number` - (Optional) The version number of the Custom Image.

## Attributes Reference

The following attributes are exported:

* `id` - The user profile Amazon Resource Name (ARN).
* `arn` - The user profile Amazon Resource Name (ARN).
* `home_efs_file_system_uid` - The ID of the user's profile in the Amazon Elastic File System (EFS) volume.


## Import

Sagemaker Code User Profiles can be imported using the `arn`, e.g.

```
$ terraform import aws_sagemaker_user_profile.test_user_profile arn:aws:sagemaker:us-west-2:123456789012:user-profile/domain-id/profile-name
```
