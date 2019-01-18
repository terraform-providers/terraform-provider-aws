package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNatGateway(t *testing.T) {
	// This is used as a portion of CIDR network addresses.
	rInt := acctest.RandIntRange(4, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNatGatewayConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_id", "id",
						"aws_nat_gateway.test", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_subnet_id", "subnet_id",
						"aws_nat_gateway.test", "subnet_id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_nat_gateway.test_by_tags", "tags.Name",
						"aws_nat_gateway.test", "tags.Name"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "state"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "allocation_id"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "network_interface_id"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "public_ip"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "private_ip"),
					resource.TestCheckNoResourceAttr("data.aws_nat_gateway.test_by_id", "attached_vpc_id"),
					resource.TestCheckResourceAttrSet("data.aws_nat_gateway.test_by_id", "tags.OtherTag"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNatGatewayConfig(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"
  tags = {
    Name = "terraform-testacc-nat-gw-data-source"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.123.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-nat-gw-data-source"
  }
}

# EIPs are not taggable
resource "aws_eip" "test" {
  vpc = true
}

# IGWs are required for an NGW to spin up; manual dependency
resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "terraform-testacc-nat-gateway-data-source-%d"
  }
}

resource "aws_nat_gateway" "test" {
  subnet_id     = "${aws_subnet.test.id}"
  allocation_id = "${aws_eip.test.id}"

  tags = {
    Name = "terraform-testacc-nat-gw-data-source-%d"
    OtherTag = "some-value"
  }

  depends_on = ["aws_internet_gateway.test"]
}

data "aws_nat_gateway" "test_by_id" {
  id = "${aws_nat_gateway.test.id}"
}

data "aws_nat_gateway" "test_by_subnet_id" {
  subnet_id = "${aws_nat_gateway.test.subnet_id}"
}

data "aws_nat_gateway" "test_by_tags" {
  tags = {
    Name = "${aws_nat_gateway.test.tags["Name"]}"
  }
}

`, rInt, rInt, rInt, rInt)
}
