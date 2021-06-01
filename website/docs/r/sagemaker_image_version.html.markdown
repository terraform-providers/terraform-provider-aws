---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_image_version"
description: |-
  Provides a Sagemaker Image Version resource.
---

# Resource: aws_sagemaker_image_version

Provides a Sagemaker Image Version resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = "012345678912.dkr.ecr.us-west-2.amazonaws.com/image:latest"
}
```

## Argument Reference

The following arguments are supported:

* `image_name` - (Required) The name of the image. Must be unique to your account.
* `base_image` - (Required) The registry path of the container image on which this image version is based.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the Image.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Image Version.
* `image_arn`- The Amazon Resource Name (ARN) of the image the version is based on.
* `container_image` - The registry path of the container image that contains this image version.

## Import

Sagemaker Image Versions can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_image_version.test_image my-code-repo
```
