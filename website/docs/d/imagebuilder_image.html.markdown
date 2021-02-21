---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image"
description: |-
    Provides details about an Image Builder Image
---

# Data Source: aws_imagebuilder_image

Provides details about an Image Builder Image.

## Example Usage

### Latest

```hcl
data "aws_imagebuilder_image" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:image/amazon-linux-2-x86/x.x.x"
}
```

## Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the image. The suffix can either be specified with wildcards (`x.x.x`) to fetch the latest build version or a full build version (e.g. `2020.11.26/1`) to fetch an exact version.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `build_version_arn` - Build version Amazon Resource Name (ARN) of the image. This will always have the `#.#.#/#` suffix.
* `date_created` - Date the image was created.
* `distribution_configuration_arn` - Amazon Resource Name (ARN) of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - Whether additional information about the image being created is collected.
* `image_recipe_arn` - Amazon Resource Name (ARN) of the Image Builder Infrastructure Recipe.
* `image_tests_configuration` - List of an object with image tests configuration.
    * `image_tests_enabled` - Whether image tests are enabled.
    * `timeout_minutes` - Number of minutes before image tests time out.
* `infrastructure_configuration_arn` - Amazon Resource Name (ARN) of the Image Builder Infrastructure Configuration.
* `name` - Name of the image.
* `platform` - Platform of the image.
* `os_version` - Operating System version of the image.
* `output_resources` - List of objects with resources created by the image.
    * `amis` - Set of objects with each Amazon Machine Image (AMI) created.
        * `account_id` - Account identifier of the AMI.
        * `description` - Description of the AMI.
        * `image` - Identifier of the AMI.
        * `name` - Name of the AMI.
        * `region` - Region of the AMI.
* `tags` - Key-value map of resource tags for the image.
* `version` - Version of the image.
