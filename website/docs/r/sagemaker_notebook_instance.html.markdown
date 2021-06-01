---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_notebook_instance"
description: |-
  Provides a Sagemaker Notebook Instance resource.
---

# Resource: aws_sagemaker_notebook_instance

Provides a Sagemaker Notebook Instance resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_notebook_instance" "ni" {
  name          = "my-notebook-instance"
  role_arn      = aws_iam_role.role.arn
  instance_type = "ml.t2.medium"

  tags = {
    Name = "foo"
  }
}
```

### Code repository usage

```hcl
resource "aws_sagemaker_code_repository" "example" {
  code_repository_name = "my-notebook-instance-code-repo"

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}

resource "aws_sagemaker_notebook_instance" "ni" {
  name                    = "my-notebook-instance"
  role_arn                = aws_iam_role.role.arn
  instance_type           = "ml.t2.medium"
  default_code_repository = aws_sagemaker_code_repository.example.code_repository_name

  tags = {
    Name = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the notebook instance (must be unique).
* `role_arn` - (Required) The ARN of the IAM role to be used by the notebook instance which allows SageMaker to call other services on your behalf.
* `instance_type` - (Required) The name of ML compute instance type.
* `volume_size` - (Optional) The size, in GB, of the ML storage volume to attach to the notebook instance. The default value is 5 GB.
* `subnet_id` - (Optional) The VPC subnet ID.
* `security_groups` - (Optional) The associated security groups.
* `kms_key_id` - (Optional) The AWS Key Management Service (AWS KMS) key that Amazon SageMaker uses to encrypt the model artifacts at rest using Amazon S3 server-side encryption.
* `lifecycle_config_name` - (Optional) The name of a lifecycle configuration to associate with the notebook instance.
* `root_access` - (Optional) Whether root access is `Enabled` or `Disabled` for users of the notebook instance. The default value is `Enabled`.
* `direct_internet_access` - (Optional) Set to `Disabled` to disable internet access to notebook. Requires `security_groups` and `subnet_id` to be set. Supported values: `Enabled` (Default) or `Disabled`. If set to `Disabled`, the notebook instance will be able to access resources only in your VPC, and will not be able to connect to Amazon SageMaker training and endpoint services unless your configure a NAT Gateway in your VPC.
* `additional_code_repositories` - (Optional) An array of up to three Git repositories to associate with the notebook instance.
 These can be either the names of Git repositories stored as resources in your account, or the URL of Git repositories in [AWS CodeCommit](https://docs.aws.amazon.com/codecommit/latest/userguide/welcome.html) or in any other Git repository. These repositories are cloned at the same level as the default repository of your notebook instance.
* `default_code_repository` - (Optional) The Git repository associated with the notebook instance as its default code repository. This can be either the name of a Git repository stored as a resource in your account, or the URL of a Git repository in [AWS CodeCommit](https://docs.aws.amazon.com/codecommit/latest/userguide/welcome.html) or in any other Git repository.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the notebook instance.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this notebook instance.
* `url` - The URL that you use to connect to the Jupyter notebook that is running in your notebook instance.
* `network_interface_id` - The network interface ID that Amazon SageMaker created at the time of creating the instance. Only available when setting `subnet_id`.

## Import

Sagemaker Notebook Instances can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_notebook_instance.test_notebook_instance my-notebook-instance
```
