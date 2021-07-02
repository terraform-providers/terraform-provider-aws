---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_device_fleet"
description: |-
  Provides a Sagemaker Device Fleet resource.
---

# Resource: aws_sagemaker_device_fleet

Provides a Sagemaker Device Fleet resource.

## Example Usage

### Basic usage

```terraform
resource "aws_sagemaker_device_fleet" "example" {
  device_fleet_name = "example"
  role_arn          = aws_iam_role.test.arn

  output_config {
    s3_output_location = "s3://${aws_s3_bucket.test.bucket}/prefix/"
  }
}
```

## Argument Reference

The following arguments are supported:

* `device_fleet_name` - (Required) The name of the Device Fleet (must be unique).
* `role_arn` - (Required) The Amazon Resource Name (ARN) that has access to AWS Internet of Things (IoT).
* `output_config` - (Required) Specifies details about the repository. see [Output Config](#output-config) details below.
* `description` - (Optional) A description of the fleet.
* `enable_iot_role_alias` - (Optional) Whether to create an AWS IoT Role Alias during device fleet creation. The name of the role alias generated will match this pattern: "SageMakerEdge-{DeviceFleetName}".
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Output Config

* `s3_output_location` - (Required) The Amazon Simple Storage (S3) bucker URI.
* `kms_key_id` - (Optional) The AWS Key Management Service (AWS KMS) key that Amazon SageMaker uses to encrypt data on the storage volume after compilation job. If you don't provide a KMS key ID, Amazon SageMaker uses the default KMS key for Amazon S3 for your role's account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the Device Fleet.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Device Fleet.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Sagemaker Device Fleets can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_device_fleet.example my-fleet
```
