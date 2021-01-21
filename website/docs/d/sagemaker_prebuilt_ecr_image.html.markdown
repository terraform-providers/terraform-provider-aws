---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_prebuilt_ecr_image"
description: |-
  Get information about prebuilt Amazon SageMaker Docker images.
---

# Data Source: aws_sagemaker_prebuilt_ecr_image

Get information about prebuilt Amazon SageMaker Docker images.

~> **NOTE:** The AWS provider creates a validly constructed `registry_path` but does not verify that the `registry_path` corresponds to an existing image. For example, using a `registry_path` containing an `image_tag` that does not correspond to a Docker image in the ECR repository, will result in an error.

## Example Usage

Basic usage:

```hcl
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
}
```

## Argument Reference

The following arguments are supported:

* `repository_name` - (Required) The name of the repository, which is generally the algorithm or library. Values include `blazingtext`, `factorization-machines`, `forecasting-deepar`, `image-classification`, `ipinsights`, `kmeans`, `knn`, `lda`, `linear-learner`, `mxnet-inference-eia`, `mxnet-inference`, `mxnet-training`, `ntm`, `object-detection`, `object2vec`, `pca`, `pytorch-inference-eia`, `pytorch-inference`, `pytorch-training`, `randomcutforest`, `sagemaker-scikit-learn`, `sagemaker-sparkml-serving`, `sagemaker-xgboost`, `semantic-segmentation`, `seq2seq`, `tensorflow-inference-eia`, `tensorflow-inference`, and `tensorflow-training`.
* `dns_suffix` - (Optional) The DNS suffix to use in the registry path. If not specified, the AWS provider sets it to the DNS suffix for the current region.
* `image_tag` - (Optional) The image tag for the Docker image. If not specified, the AWS provider sets the value to `1`, which for many repositories indicates the latest version. Some repositories, such as XGBoost, do not support `1` or `latest` and specific version must be used.
* `region` (Optional) - The region to use in the registry path. If not specified, the AWS provider sets it to the current region.

## Attributes Reference

The following attributes are exported:

* `registry_id` - The account ID containing the image. For example, `469771592824`.
* `registry_path` - The Docker image URL. For example, `341280168497.dkr.ecr.ca-central-1.amazonaws.com/sagemaker-sparkml-serving:2.4`.
