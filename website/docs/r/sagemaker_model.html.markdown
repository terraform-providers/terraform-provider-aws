---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_model"
description: |-
  Provides a SageMaker model resource.
---

# Resource: aws_sagemaker_model

Provides a SageMaker model resource.

## Example Usage

Basic usage:

```hcl
resource "aws_sagemaker_model" "example" {
  name               = "my-model"
  execution_role_arn = aws_iam_role.example.arn

  primary_container {
    image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
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

* `name` - (Optional) The name of the model (must be unique). If omitted, Terraform will assign a random, unique name.
* `primary_container` - (Optional) The primary docker image containing inference code that is used when the model is deployed for predictions.  If not specified, the `container` argument is required. Fields are documented below.
* `execution_role_arn` - (Required) A role that SageMaker can assume to access model artifacts and docker images for deployment.
* `container` (Optional) -  Specifies containers in the inference pipeline. If not specified, the `primary_container` argument is required. Fields are documented below.
* `enable_network_isolation` (Optional) - Isolates the model container. No inbound or outbound network calls can be made to or from the model container.
* `vpc_config` (Optional) - Specifies the VPC that you want your model to connect to. VpcConfig is used in hosting services and in batch transform.
* `tags` - (Optional) A map of tags to assign to the resource.

The `primary_container` and `container` block both support:

* `image` - (Required) The registry path where the inference code image is stored in Amazon ECR.
* `mode` - (Optional) The container hosts value `SingleModel/MultiModel`. The default value is `SingleModel`.
* `model_data_url` - (Optional) The URL for the S3 location where model artifacts are stored.
* `container_hostname` - (Optional) The DNS host name for the container.
* `environment` - (Optional) Environment variables for the Docker container.
   A list of key value pairs.
* `image_config` - (Optional) Specifies whether the model container is in Amazon ECR or a private Docker registry accessible from your Amazon Virtual Private Cloud (VPC). For more information see [Using a Private Docker Registry for Real-Time Inference Containers](https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-containers-inference-private.html). see [Image Config](#image-config).

### Image Config

* `repository_access_mode` - (Required) Specifies whether the model container is in Amazon ECR or a private Docker registry accessible from your Amazon Virtual Private Cloud (VPC). Allowed values are: `Platform` and `Vpc`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the model.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this model.

## Import

Models can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_model.test_model model-foo
```
