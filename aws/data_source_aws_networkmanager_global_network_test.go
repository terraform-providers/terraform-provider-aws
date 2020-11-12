package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSNetworkManagerGlobalNetwork_basic(t *testing.T) {
	dataSourceByIdName := "data.aws_networkmanager_global_network.test_by_id"
	dataSourceByTagsName := "data.aws_networkmanager_global_network.test_by_tags"
	resourceName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerGlobalNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkManagerGlobalNetworkConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "tags.OtherTag", resourceName, "tags.OtherTag"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkManagerGlobalNetworkConfig() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = "test"

  tags = {
    Name     = "terraform-testacc-global-network-%d"
    OtherTag = "some-value"
  }
}

resource "aws_networkmanager_global_network" "test2" {
  description = "test2"

  tags = {
    Name     = "terraform-testacc-global-network2"
  }
}

data "aws_networkmanager_global_network" "test_by_id" {
  id = aws_networkmanager_global_network.test.id
}

data "aws_networkmanager_global_network" "test_by_tags" {
    tags = {
	Name = aws_networkmanager_global_network.test.tags["Name"]
  }
}
`, acctest.RandInt())
}
