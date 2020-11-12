package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSNetworkManagerDevice_basic(t *testing.T) {
	dataSourceName := "data.aws_networkmanager_device.test"
	dataSourceByIdName := "data.aws_networkmanager_device.test_by_id"
	dataSourceBySiteIdName := "data.aws_networkmanager_device.test_by_site_id"
	dataSourceByTagsName := "data.aws_networkmanager_device.test_by_tags"
	resourceName := "aws_networkmanager_device.test"
	resourceSiteName := "aws_networkmanager_site.test"
	resourceGlobalNetworkName := "aws_networkmanager_global_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkManagerDeviceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceGlobalNetworkName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "model", resourceName, "model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "serial_number", resourceName, "serial_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site_id", resourceSiteName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vendor", resourceName, "vendor"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.OtherTag", resourceName, "tags.OtherTag"),
					resource.TestCheckResourceAttr(dataSourceName, "location.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.address", ""),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.latitude", "18.0029784"),
					resource.TestCheckResourceAttr(dataSourceName, "location.0.longitude", "-76.7897987"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceBySiteIdName, "site_id", resourceSiteName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByTagsName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkManagerDeviceConfig() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = "test"
}

resource "aws_networkmanager_site" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id

}

resource "aws_networkmanager_device" "test" {
 description       = "test"
 global_network_id = aws_networkmanager_global_network.test.id
 site_id           = aws_networkmanager_site.test.id
 model             = "abc"
 serial_number     = "123"
 type              = "office device"
 vendor            = "company"

  location {
   latitude  = "18.0029784"	
   longitude = "-76.7897987"
  }

  tags = {
    Name     = "terraform-testacc-site-%d"
    OtherTag = "some-value"
  }
}

data "aws_networkmanager_device" "test" {
  global_network_id = aws_networkmanager_device.test.global_network_id
}

data "aws_networkmanager_device" "test_by_id" {
  global_network_id = aws_networkmanager_global_network.test.id
  id                = aws_networkmanager_device.test.id
}

data "aws_networkmanager_device" "test_by_site_id" {
  global_network_id = aws_networkmanager_device.test.global_network_id
  site_id           = aws_networkmanager_site.test.id
}

data "aws_networkmanager_device" "test_by_tags" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
	Name = aws_networkmanager_device.test.tags["Name"]
  }
}
`, acctest.RandInt())
}
