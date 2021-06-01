package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsImageBuilderImagePipelineDataSource_Arn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_imagebuilder_image_pipeline.test"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderImagePipelineDataSourceConfigArn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_last_run", resourceName, "date_last_run"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_next_run", resourceName, "date_next_run"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_updated", resourceName, "date_updated"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_configuration_arn", resourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enhanced_image_metadata_enabled", resourceName, "enhanced_image_metadata_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_recipe_arn", resourceName, "image_recipe_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_tests_configuration.#", resourceName, "image_tests_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "infrastructure_configuration_arn", resourceName, "infrastructure_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, "schedule.#", resourceName, "schedule.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccAwsImageBuilderImagePipelineDataSourceConfigArn(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}

resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}

data "aws_imagebuilder_image_pipeline" "test" {
  arn = aws_imagebuilder_image_pipeline.test.arn
}
`, rName)
}
