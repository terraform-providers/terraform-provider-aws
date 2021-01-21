---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_domain"
description: |-
  Provides a Sagemaker Domain resource.
---

# Resource: aws_sagemaker_domain

Provides a Sagemaker Domain resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_domain" "example" {
  domain_name = "example"
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}

resource "aws_iam_role" "example" {
  name               = "example"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The domain name.
* `auth_mode` - (Required) The mode of authentication that members use to access the domain. Valid values are `IAM` and `SSO`.
* `vpc_id` - (Required) The ID of the Amazon Virtual Private Cloud (VPC) that Studio uses for communication.
* `subnet_ids` - (Required) The VPC subnets that Studio uses for communication.
* `default_user_settings` - (Required) The default user settings. See [Default User Settings](#default-user-settings) below.
* `kms_key_id` - (Optional) The AWS KMS customer managed CMK used to encrypt the EFS volume attached to the domain.
* `app_network_access_type` - (Optional) Specifies the VPC used for non-EFS traffic. The default value is `PublicInternetOnly`. Valid values are `PublicInternetOnly` and `VpcOnly`.
* `tags` - (Optional) A map of tags to assign to the resource.

### Default User Settings

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

* `id` - The ID of the Domain.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Domain.
* `url` - The domain's URL.
* `single_sign_on_managed_application_instance_id` - The SSO managed application instance ID.
* `home_efs_file_system_id` - The ID of the Amazon Elastic File System (EFS) managed by this Domain.


## Import

Sagemaker Code Domains can be imported using the `id`, e.g.

```
$ terraform import aws_sagemaker_domain.test_domain d-8jgsjtilstu8
```
