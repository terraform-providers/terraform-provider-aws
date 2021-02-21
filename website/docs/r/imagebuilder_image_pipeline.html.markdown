---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_pipeline"
description: |-
    Manages an Image Builder Image Pipeline
---

# Resource: aws_imagebuilder_image_pipeline

Manages an Image Builder Image Pipeline.

## Example Usage

```hcl
resource "aws_imagebuilder_image_pipeline" "example" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.example.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.example.arn
  name                             = "example"

  schedule {
    schedule_expression = "cron(0 0 * * ? *)"
  }
}
```

## Argument Reference

The following arguments are required:

* `image_recipe_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Infrastructure Recipe.
* `infrastructure_configuration_arn` - (Required) Amazon Resource Name (ARN) of the Image Builder Infrastructure Configuration.
* `name` - (Required) Name of the image pipeline.

The following arguments are optional:

* `description` - (Optional) Description of the image pipeline.
* `distribution_configuration_arn` - (Optional) Amazon Resource Name (ARN) of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - (Optional) Whether additional information about the image being created is collected. Defaults to `true`.
* `image_tests_configuration` - (Optional) Configuration block with image tests configuration. Detailed below.
* `schedule` - (Optional) Configuration block with schedule settings. Detailed below.
* `status` - (Optional) Status of the image pipeline. Valid values are `DISABLED` and `ENABLED`. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags for the image pipeline.

### image_tests_configuration

The following arguments are optional:

* `image_tests_enabled` - (Optional) Whether image tests are enabled. Defaults to `true`.
* `timeout_minutes` - (Optional) Number of minutes before image tests time out. Valid values are between `60` and `1440`. Defaults to `720`.

### schedule

The following arguments are required:

* `schedule_expression` - (Required) Cron expression of how often the pipeline start condition is evaluated. For example, `cron(0 0 * * ? *)` is evaluated every day at midnight UTC. Configurations using the five field syntax that was previously accepted by the API, such as `cron(0 0 * * *)`, must be updated to the six field syntax. For more information, see the [Image Builder User Guide](https://docs.aws.amazon.com/imagebuilder/latest/userguide/cron-expressions.html).

The following arguments are optional:

* `pipeline_execution_start_condition` - (Optional) Condition when the pipeline should trigger a new image build. Valid values are `EXPRESSION_MATCH_AND_DEPENDENCY_UPDATES_AVAILABLE` and `EXPRESSION_MATCH_ONLY`. Defaults to `EXPRESSION_MATCH_AND_DEPENDENCY_UPDATES_AVAILABLE`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the image pipeline.
* `date_created` - Date the image pipeline was created.
* `date_last_run` - Date the image pipeline was last run.
* `date_next_run` - Date the image pipeline will run next.
* `date_updated` - Date the image pipeline was updated.
* `platform` - Platform of the image pipeline.

## Import

`aws_imagebuilder_image_pipeline` resources can be imported using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_imagebuilder_image_pipeline.example arn:aws:imagebuilder:us-east-1:123456789012:image-pipeline/example
```
