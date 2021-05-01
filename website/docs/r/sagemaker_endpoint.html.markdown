---
subcategory: "Sagemaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_endpoint"
description: |-
  Provides a SageMaker Endpoint resource.
---

# Resource: aws_sagemaker_endpoint

Provides a SageMaker Endpoint resource.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_endpoint" "e" {
  name                 = "my-endpoint"
  endpoint_config_name = aws_sagemaker_endpoint_configuration.ec.name

  tags = {
    Name = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `endpoint_config_name` - (Required) The name of the endpoint configuration to use.
* `name` - (Optional) The name of the endpoint. If omitted, Terraform will assign a random, unique name.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this endpoint.
* `name` - The name of the endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Endpoints can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_endpoint.test_endpoint my-endpoint
```
