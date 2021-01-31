---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_app_image_config"
description: |-
  Provides a Sagemaker App Image Config resource.
---

# Resource: aws_sagemaker_app_image_config

Provides a Sagemaker App Image Config resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = "example"

  kernel_gateway_image_config {
    kernel_spec {
      name = "example"
    }
  }
}
```

### Default File System Config

```hcl
resource "aws_sagemaker_app_image_config" "test" {
  app_image_config_name = "example"

  kernel_gateway_image_config {
    kernel_spec {
      name = "example"
    }

    file_system_config {}
  }
}
```

## Argument Reference

The following arguments are supported:

* `app_image_config_name` - (Required) The name of the App Image Config.
* `kernel_gateway_image_config` - (Optional) The configuration for the file system and kernels in a SageMaker image running as a KernelGateway app. See [Kernel Gateway Image Config](#kernel-gateway-image-config) details below.
* `tags` - (Optional) A map of tags to assign to the resource.

### Kernel Gateway Image Config

* `file_system_config` - (Optional) The URL where the Git repository is located. See [File System Config](#file-system-config) details below.
* `kernel_spec` - (Required) The default branch for the Git repository. See [Kernel Spec](#kernel-spec) details below.

#### File System Config

* `default_gid` - (Optional) The default POSIX group ID (GID). If not specified, defaults to `100`. Valid values are `0` and `100`.
* `default_uid` - (Optional) The default POSIX user ID (UID). If not specified, defaults to `1000`. Valid values are `0` and `1000`.
* `mount_path` - (Optional) The path within the image to mount the user's EFS home directory. The directory should be empty. If not specified, defaults to `/home/sagemaker-user`.

~> **Note:** When specifying `default_gid` and `default_uid`, Valid value pairs are [`0`, `0`] and [`100`, `1000`].

#### Kernel Spec

* `name` - (Required) The name of the kernel.
* `display_name` - (Optional) The display name of the kernel.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the App Image Config.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this App Image Config.

## Import

Sagemaker App Image Configs can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_app_image_config.example example
```
