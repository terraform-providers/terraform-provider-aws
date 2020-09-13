package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSMskConfigurationDataSource_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_msk_configuration.test"
	resourceName := "aws_msk_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskConfigurationDataSourceConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "kafka_versions.#", dataSourceName, "kafka_versions.#"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision", dataSourceName, "latest_revision"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "server_properties", dataSourceName, "server_properties"),
				),
			},
		},
	})
}

func testAccMskConfigurationDataSourceConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_msk_configuration" "test" {
  kafka_versions = ["2.1.0"]
  name           = %[1]q

  server_properties = <<PROPERTIES
auto.create.topics.enable = true
delete.topic.enable = true
PROPERTIES
}

data "aws_msk_configuration" "test" {
  name = aws_msk_configuration.test.name
}
`, rName)
}
