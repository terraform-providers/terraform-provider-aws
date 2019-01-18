package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsSubnetIDs(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetIDsConfig(rInt),
			},
			{
				Config: testAccDataSourceAwsSubnetIDsConfigWithDataSource(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnet_ids.selected", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_subnet_ids.private", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSubnetIDs_filter(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	rName := "data.aws_subnet_ids.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetIDs_filter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rName, "ids.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSubnetIDsConfigWithDataSource(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-ids-data-source"
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.123.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-public-a"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.125.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-private-a"
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.126.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-private-b"
    Tier = "Private"
  }
}

data "aws_subnet_ids" "selected" {
  vpc_id = "${aws_vpc.test.id}"
}

data "aws_subnet_ids" "private" {
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Tier = "Private"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccDataSourceAwsSubnetIDsConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-ids-data-source"
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.123.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-public-a"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.125.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-private-a"
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "172.%d.126.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-subnet-ids-data-source-private-b"
    Tier = "Private"
  }
}
`, rInt, rInt, rInt, rInt)
}

func testAccDataSourceAwsSubnetIDs_filter(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"
  tags = {
    Name = "terraform-testacc-subnet-ids-data-source"
  }
}

resource "aws_subnet" "test_a_one" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.%d.1.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_subnet" "test_a_two" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.%d.2.0/24"
  availability_zone = "us-west-2a"
}

resource "aws_subnet" "test_b" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.%d.3.0/24"
  availability_zone = "us-west-2b"
}

data "aws_subnet_ids" "test" {
  vpc_id = "${aws_subnet.test_a_two.vpc_id}"
  filter {
    name = "availabilityZone"
    values = ["${aws_subnet.test_a_one.availability_zone}"]
  }
}
`, rInt, rInt, rInt, rInt)
}
