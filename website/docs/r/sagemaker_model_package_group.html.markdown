---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_model_package_group"
description: |-
  Provides a Sagemaker Model Package Group resource.
---

# Resource: aws_sagemaker_model_package_group

Provides a Sagemaker Model Package Group resource.

## Example Usage

### Basic usage

```hcl
resource "aws_sagemaker_model_package_group" "example" {
  model_package_group_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `model_package_group_name` - (Required) The name of the model group.
* `model_package_group_description` - (Optional) A description for the model group.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The name of the Model Package Group.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this Model Package Group.

## Import

Sagemaker Code Model Package Groups can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_model_package_group.test_model_package_group my-code-repo
```
