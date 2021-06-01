---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_endpoint_configuration"
description: |-
  Provides a SageMaker Endpoint Configuration resource.
---

# Resource: aws_sagemaker_endpoint_configuration

Provides a SageMaker endpoint configuration resource.

## Example Usage


Basic usage:

```hcl
resource "aws_sagemaker_endpoint_configuration" "ec" {
  name = "my-endpoint-config"

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.m.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  tags = {
    Name = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `production_variants` - (Required) Fields are documented below.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt data on the storage volume attached to the ML compute instance that hosts the endpoint.
* `name` - (Optional) The name of the endpoint configuration. If omitted, Terraform will assign a random, unique name.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `data_capture_config` - (Optional) Specifies the parameters to capture input/output of Sagemaker models endpoints. Fields are documented below.

The `production_variants` block supports:

* `initial_instance_count` - (Required) Initial number of instances used for auto-scaling.
* `instance_type` (Required) - The type of instance to start.
* `accelerator_type` (Optional) - The size of the Elastic Inference (EI) instance to use for the production variant.
* `initial_variant_weight` (Optional) - Determines initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, it defaults to 1.0.
* `model_name` - (Required) The name of the model to use.
* `variant_name` - (Optional) The name of the variant. If omitted, Terraform will assign a random, unique name.

The `data_capture_config` block supports:

* `initial_sampling_percentage` - (Required) Portion of data to capture. Should be between 0 and 100.
* `destination_s3_uri` - (Required) The URL for S3 location where the captured data is stored.
* `capture_options` - (Required) Specifies what data to capture. Fields are documented below.
* `kms_key_id` - (Optional) Amazon Resource Name (ARN) of a AWS Key Management Service key that Amazon SageMaker uses to encrypt the captured data on Amazon S3.
* `enable_capture` - (Optional) Flag to enable data capture. Defaults to `false`.
* `capture_content_type_header` - (Optional) The content type headers to capture. Fields are documented below.

The `capture_options` block supports:

* `capture_mode` - (Required) Specifies the data to be captured. Should be one of `Input` or `Output`.

The `capture_content_type_header` block supports:

* `csv_content_types` - (Optional) The CSV content type headers to capture.
* `json_content_types` - (Optional) The JSON content type headers to capture.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint configuration.
* `name` - The name of the endpoint configuration.

## Import

Endpoint configurations can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_endpoint_configuration.test_endpoint_config endpoint-config-foo
```
