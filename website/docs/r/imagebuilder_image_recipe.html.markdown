---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_recipe"
description: |-
    Manage an Image Builder Image Recipe
---

# Resource: aws_imagebuilder_image_recipe

Manages an Image Builder Image Recipe.

## Example Usage

```hcl
resource "aws_imagebuilder_image_recipe" "example" {
  block_device_mapping {
    device_name = "/dev/xvdb"

    ebs {
      delete_on_termination = true
      volume_size           = 100
      volume_type           = "gp2"
    }
  }

  component {
    component_arn = aws_imagebuilder_component.example.arn
  }

  name         = "example"
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
```

## Argument Reference

The following arguments are required:

* `component` - (Required) Ordered configuration block(s) with components for the image recipe. Detailed below.
* `name` - (Required) Name of the image recipe.
* `parent_image` - (Required) Platform of the image recipe.
* `version` - (Required) Version of the image recipe.

The following attributes are optional:

* `block_device_mapping` - (Optional) Configuration block(s) with block device mappings for the the image recipe. Detailed below.
* `description` - (Optional) Description of the image recipe.
* `tags` - (Optional) Key-value map of resource tags for the image recipe.
* `working_directory` - (Optional) The working directory to be used during build and test workflows.

### block_device_mapping

The following arguments are optional:

* `device_name` - (Optional) Name of the device. For example, `/dev/sda` or `/dev/xvdb`.
* `ebs` - (Optional) Configuration block with Elastic Block Storage (EBS) block device mapping settings. Detailed below.
* `no_device` - (Optional) Set to `true` to remove a mapping from the parent image.
* `virtual_name` - (Optional) Virtual device name. For example, `ephemeral0`. Instance store volumes are numbered starting from 0.

#### ebs

The following arguments are optional:

* `delete_on_termination` - (Optional) Whether to delete the volume on termination. Defaults to unset, which is the value inherited from the parent image.
* `encrypted` - (Optional) Whether to encrypt the volume. Defaults to unset, which is the value inherited from the parent image.
* `iops` - (Optional) Number of Input/Output (I/O) operations per second to provision for an `io1` or `io2` volume.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of the Key Management Service (KMS) Key for encryption.
* `snapshot_id` - (Optional) Identifier of the EC2 Volume Snapshot.
* `volume_size` - (Optional) Size of the volume, in GiB.
* `volume_type` - (Optional) Type of the volume. For example, `gp2` or `io2`.

### component

The following arguments are required:

* `component_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Component to associate.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - (Required) Amazon Resource Name (ARN) of the image recipe.
* `date_created` - Date the image recipe was created.
* `owner` - Owner of the image recipe.
* `platform` - Platform of the image recipe.

## Import

`aws_imagebuilder_image_recipe` resources can be imported by using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_imagebuilder_image_recipe.example arn:aws:imagebuilder:us-east-1:123456789012:image-recipe/example/1.0.0
```
