---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_recipe"
description: |-
    Provides details about an Image Builder Image Recipe
---

# Data Source: aws_imagebuilder_image_recipe

Provides details about an Image Builder Image Recipe.

## Example Usage

```hcl
data "aws_imagebuilder_image_recipe" "example" {
  arn = "arn:aws:imagebuilder:us-east-1:aws:image-recipe/example/1.0.0"
}
```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Amazon Resource Name (ARN) of the image recipe.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `block_device_mapping` - Set of objects with block device mappings for the the image recipe.
    * `device_name` - Name of the device. For example, `/dev/sda` or `/dev/xvdb`.
    * `ebs` - Single list of object with Elastic Block Storage (EBS) block device mapping settings.
        * `delete_on_termination` - Whether to delete the volume on termination. Defaults to unset, which is the value inherited from the parent image.
        * `encrypted` - Whether to encrypt the volume. Defaults to unset, which is the value inherited from the parent image.
        * `iops` - Number of Input/Output (I/O) operations per second to provision for an `io1` or `io2` volume.
        * `kms_key_id` - Amazon Resource Name (ARN) of the Key Management Service (KMS) Key for encryption.
        * `snapshot_id` - Identifier of the EC2 Volume Snapshot.
        * `volume_size` - Size of the volume, in GiB.
        * `volume_type` - Type of the volume. For example, `gp2` or `io2`.
    * `no_device` - Whether to remove a mapping from the parent image.
    * `virtual_name` - Virtual device name. For example, `ephemeral0`. Instance store volumes are numbered starting from 0.
* `component` - List of objects with components for the image recipe.
    * `component_arn` - Amazon Resource Name (ARN) of the Image Builder Component.
* `date_created` - Date the image recipe was created.
* `description` - Description of the image recipe.
* `name` - Name of the image recipe.
* `owner` - Owner of the image recipe.
* `parent_image` - Platform of the image recipe.
* `platform` - Platform of the image recipe.
* `tags` - Key-value map of resource tags for the image recipe.
* `version` - Version of the image recipe.
* `working_directory` - The working directory used during build and test workflows.
