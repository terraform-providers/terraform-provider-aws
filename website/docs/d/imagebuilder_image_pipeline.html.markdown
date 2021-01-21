---
subcategory: "Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_pipeline"
description: |-
    Provides details about an Image Builder Image Pipeline
---

# Data Source: aws_imagebuilder_image_pipeline

Provides details about an Image Builder Image Pipeline.

## Example Usage

```hcl
data "aws_imagebuilder_image_pipeline" "example" {
  arn = "arn:aws:imagebuilder:us-west-2:aws:image-pipeline/example"
}
```

## Argument Reference

* `arn` - (Required) Amazon Resource Name (ARN) of the image pipeline.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `date_created` - Date the image pipeline was created.
* `date_last_run` - Date the image pipeline was last run.
* `date_next_run` - Date the image pipeline will run next.
* `date_updated` - Date the image pipeline was updated.
* `description` - Description of the image pipeline.
* `distribution_configuration_arn` - Amazon Resource Name (ARN) of the Image Builder Distribution Configuration.
* `enhanced_image_metadata_enabled` - Whether additional information about the image being created is collected.
* `image_recipe_arn` - Amazon Resource Name (ARN) of the Image Builder Infrastructure Recipe.
* `image_tests_configuration` - List of an object with image tests configuration.
    * `image_tests_enabled` - Whether image tests are enabled.
    * `timeout_minutes` - Number of minutes before image tests time out.
* `infrastructure_configuration_arn` - Amazon Resource Name (ARN) of the Image Builder Infrastructure Configuration.
* `name` - Name of the image pipeline.
* `platform` - Platform of the image pipeline.
* `schedule` - List of an object with schedule settings.
    * `pipeline_execution_start_condition` - Condition when the pipeline should trigger a new image build.
    * `schedule_expression` - Cron expression of how often the pipeline start condition is evaluated.
* `status` - Status of the image pipeline.
* `tags` - Key-value map of resource tags for the image pipeline.
