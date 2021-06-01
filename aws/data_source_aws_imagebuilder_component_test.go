package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsImageBuilderComponentDataSource_Arn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_imagebuilder_component.test"
	resourceName := "aws_imagebuilder_component.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsImageBuilderComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsImageBuilderComponentDataSourceConfigBuildVersionArn(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "change_description", resourceName, "change_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data", resourceName, "data"),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner", resourceName, "owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, "supported_os_versions.#", resourceName, "supported_os_versions.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccAwsImageBuilderComponentDataSourceConfigBuildVersionArn(rName string) string {
	return fmt.Sprintf(`
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

data "aws_imagebuilder_component" "test" {
  arn = aws_imagebuilder_component.test.arn
}
`, rName)
}
