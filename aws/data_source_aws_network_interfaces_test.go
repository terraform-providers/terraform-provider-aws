package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNetworkInterfaces_Filter(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterfacesConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interfaces.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkInterfaces_Tags(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterfacesConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interfaces.test", "ids.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkInterfacesConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-eni-data-source-basic-%s"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "terraform-testacc-eni-data-source-basic-%s"
  }
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
}

resource "aws_network_interface" "test1" {
  subnet_id = "${aws_subnet.test.id}"
  tags = {
  	Name = "${aws_vpc.test.tags.Name}"
  }
}

`, rName, rName)
}

func testAccDataSourceAwsNetworkInterfacesConfig_Filter(rName string) string {
	return testAccDataSourceAwsNetworkInterfacesConfig_Base(rName) + `
data "aws_network_interfaces" "test" {
  filter {
    name   = "subnet-id"
    values = ["${aws_network_interface.test.subnet_id}", "${aws_network_interface.test1.subnet_id}"]
  }
}
`
}

func testAccDataSourceAwsNetworkInterfacesConfig_Tags(rName string) string {
	return testAccDataSourceAwsNetworkInterfacesConfig_Base(rName) + `
data "aws_network_interfaces" "test" {
  tags = {
    Name = "${aws_network_interface.test1.tags.Name}"
  }
}
`
}
