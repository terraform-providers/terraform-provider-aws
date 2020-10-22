---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecr_image"
description: |-
    Provides details about an ECR Image
---

# Data Source: aws_ecr_image

The ECR Image data source allows the details of an image with a particular tag or digest to be retrieved.

## Example Usage

```hcl
data "aws_ecr_image" "service_image" {
  repository_name = "my/service"
  image_tag       = "latest"
}
```

## Argument Reference

The following arguments are supported:

* `registry_id` - (Optional) The ID of the Registry where the repository resides.
* `repository_name` - (Required) The name of the ECR Repository.
* `image_digest` - (Optional) The sha256 digest of the image manifest. At least one of `image_digest` or `image_tag` must be specified.
* `image_tag` - (Optional) The tag associated with this image. At least one of `image_digest` or `image_tag` must be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - SHA256 digest of the image manifest.
* `image_pushed_at` - The date and time, expressed as a unix timestamp, at which the current image was pushed to the repository.
* `image_size_in_bytes` - The size, in bytes, of the image in the repository.
* `image_tags` - The list of tags associated with this image.
