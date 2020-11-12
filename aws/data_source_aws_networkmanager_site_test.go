package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSNetworkManagerSite_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_site.test"
	dataSourceByIdName := "data.aws_networkmanager_site.test_by_id"
	dataSourceByTagsName := "data.aws_networkmanager_site.test_by_tags"
	resourceName := "aws_networkmanager_site.test"
	resourceGlobalNetworkName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerSiteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkManagerSiteConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceGlobalNetworkName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.OtherTag", resourceName, "tags.OtherTag"),
					resource.TestCheckResourceAttr(dataSourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.address", ""),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.latitude", "18.0029784"),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.longitude", "-76.7897987"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkManagerSiteConfig() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id

  location {
   latitude  = "18.0029784"	
   longitude = "-76.7897987"
  }

  tags = {
    Name     = "terraform-testacc-site-%d"
    OtherTag = "some-value"
  }
}

data "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_site.test.global_network_id
}

data "aws_networkmanager_site" "test_by_id" {
  global_network_id = aws_networkmanager_global_network.test.id
  id                = aws_networkmanager_site.test.id
}

data "aws_networkmanager_site" "test_by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
	Name = aws_networkmanager_site.test.tags["Name"]
  }
}
`, acctest.RandInt())
}
