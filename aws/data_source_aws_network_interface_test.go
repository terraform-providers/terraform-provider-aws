package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNetworkInterface_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterface_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interface.test", "private_ips.#", "1"),
					resource.TestCheckResourceAttr("data.aws_network_interface.test", "security_groups.#", "1"),
					resource.TestCheckResourceAttrPair("data.aws_network_interface.test", "private_ip", "aws_network_interface.test", "private_ip"),
					resource.TestCheckResourceAttrSet("data.aws_network_interface.test", "availability_zone"),
					resource.TestCheckResourceAttrPair("data.aws_network_interface.test", "description", "aws_network_interface.test", "description"),
					resource.TestCheckResourceAttrSet("data.aws_network_interface.test", "interface_type"),
					resource.TestCheckResourceAttrPair("data.aws_network_interface.test", "private_dns_name", "aws_network_interface.test", "private_dns_name"),
					resource.TestCheckResourceAttrPair("data.aws_network_interface.test", "subnet_id", "aws_network_interface.test", "subnet_id"),
					resource.TestCheckResourceAttrSet("data.aws_network_interface.test", "vpc_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkInterface_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-eni-data-source-basic"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-eni-data-source-basic"
  }
}

resource "aws_security_group" "test" {
  name = "tf-sg-%s"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["10.0.0.50"]
  security_groups = ["${aws_security_group.test.id}"]
}

data "aws_network_interface" "test" {
  id = "${aws_network_interface.test.id}"
}
`, rName)
}

func TestAccDataSourceAwsNetworkInterface_filters(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkInterface_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_interface.test", "private_ips.#", "1"),
					resource.TestCheckResourceAttr("data.aws_network_interface.test", "security_groups.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkInterface_filters(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-eni-data-source-filters"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-eni-data-source-filters"
  }
}

resource "aws_security_group" "test" {
  name = "tf-sg-%s"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
  private_ips = ["10.0.0.60"]
  security_groups = ["${aws_security_group.test.id}"]
}

data "aws_network_interface" "test" {
  filter {
    name   = "network-interface-id"
    values = ["${aws_network_interface.test.id}"]
  }
}
`, rName)
}
