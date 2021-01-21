---
subcategory: "DataPipeline"
layout: "aws"
page_title: "AWS: aws_datapipeline_pipeline"
description: |-
  Provides a AWS DataPipeline Pipeline.
---

# Resource: aws_datapipeline_pipeline

Provides a Data Pipeline resource.

## Example Usage

```hcl
resource "aws_datapipeline_pipeline" "default" {
  name = "tf-pipeline-default"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Pipeline.
* `description` - (Optional) The description of Pipeline.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the client certificate.

## Import

`aws_datapipeline_pipeline` can be imported by using the id (Pipeline ID), e.g.

```
$ terraform import aws_datapipeline_pipeline.default df-1234567890
```
