---
subcategory: "ECR"
layout: "aws"
page_title: "AWS: aws_ecr_repository"
description: |-
  Provides an Elastic Container Registry Repository.
---

# Resource: aws_ecr_repository

Provides an Elastic Container Registry Repository.

## Example Usage

```hcl
resource "aws_ecr_repository" "foo" {
  name                 = "bar"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the repository.
* `encryption_configuration` - (Optional) Encryption configuration for the repository. See [below for schema](#encryption_configuration).
* `image_tag_mutability` - (Optional) The tag mutability setting for the repository. Must be one of: `MUTABLE` or `IMMUTABLE`. Defaults to `MUTABLE`.
* `image_scanning_configuration` - (Optional) Configuration block that defines image scanning configuration for the repository. By default, image scanning must be manually triggered. See the [ECR User Guide](https://docs.aws.amazon.com/AmazonECR/latest/userguide/image-scanning.html) for more information about image scanning.
    * `scan_on_push` - (Required) Indicates whether images are scanned after being pushed to the repository (true) or not scanned (false).
* `tags` - (Optional) A map of tags to assign to the resource.

### encryption_configuration

* `encryption_type` - (Optional) The encryption type to use for the repository. Valid values are `AES256` or `KMS`. Defaults to `AES256`.
* `kms_key` - (Optional) The ARN of the KMS key to use when `encryption_type` is `KMS`. If not specified, uses the default AWS managed key for ECR.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Full ARN of the repository.
* `registry_id` - The registry ID where the repository was created.
* `repository_url` - The URL of the repository (in the form `aws_account_id.dkr.ecr.region.amazonaws.com/repositoryName`).

## Timeouts

`aws_ecr_repository` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `delete` - (Default `20 minutes`) How long to wait for a repository to be deleted.

## Import

ECR Repositories can be imported using the `name`, e.g.

```
$ terraform import aws_ecr_repository.service test-service
```
