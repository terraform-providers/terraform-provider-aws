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
    model_name             = "${aws_sagemaker_model.m.name}"
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
  }

  tags {
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

The `production_variants` block supports:

* `initial_instance_count` - (Required) Initial number of instances used for auto-scaling.
* `instance_type` (Required) - The type of instance to start.
* `accelerator_type` (Optional) - The size of the Elastic Inference (EI) instance to use for the production variant.
* `initial_variant_weight` (Optional) - Determines initial traffic distribution among all of the models that you specify in the endpoint configuration. If unspecified, it defaults to 1.0.
* `model_name` - (Required) The name of the model to use.
* `variant_name` - (Optional) The name of the variant. If omitted, Terraform will assign a random, unique name.

## Attributes Reference

The following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint configuration.
* `name` - The name of the endpoint configuration.

## Import

Endpoint configurations can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_endpoint_configuration.test_endpoint_config endpoint-config-foo
```
