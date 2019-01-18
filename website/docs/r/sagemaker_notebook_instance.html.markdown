---
layout: "aws"
page_title: "AWS: sagemaker_notebook_instance"
sidebar_current: "docs-aws-resource-sagemaker-notebook-instance"
description: |-
  Provides a Sagemaker Notebook Instance resource.
---

# aws_sagemaker_notebook_instance

Provides a Sagemaker Notebook Instance resource.

## Example Usage

Basic usage:

```hcl
resource "aws_sagemaker_notebook_instance" "ni" {
  name = "my-notebook-instance"
  role_arn = "${aws_iam_role.role.arn}"
  instance_type = "ml.t2.medium"

  tags {
    Name = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the notebook instance (must be unique).
* `role_arn` - (Required) The ARN of the IAM role to be used by the notebook instance which allows SageMaker to call other services on your behalf.
* `instance_type` - (Required) The name of ML compute instance type.
* `subnet_id` - (Optional) The VPC subnet ID.
* `security_groups` - (Optional) The associated security groups.
* `kms_key_id` - (Optional) The AWS Key Management Service (AWS KMS) key that Amazon SageMaker uses to encrypt the model artifacts at rest using Amazon S3 server-side encryption.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the notebook instance.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this notebook instance.

## Import

Sagemaker Notebook Instances can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_notebook_instance.test_notebook_instance my-notebook-instance
```