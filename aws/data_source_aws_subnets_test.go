package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsSubnets_basic(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetsConfig(rInt),
			},
			{
				Config: testAccDataSourceAwsSubnetsConfigWithDataSource(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnets.selected", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_subnets.private", "ids.#", "2"),
					testCheckResourceAttrGreaterThanValue("data.aws_subnets.all", "ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSubnets_filter(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnets_filter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_subnets.test_vpc_id", "ids.#", "2"),
					testCheckResourceAttrGreaterThanValue("data.aws_subnets.test_no_vpc_id", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSubnetsConfigWithDataSource(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[1]d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnets-data-source"
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnets-data-source-public-a"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_public_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.124.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnets-data-source-public-b"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnets-data-source-private-a"
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-subnets-data-source-private-b"
    Tier = "Private"
  }
}

data "aws_subnets" "selected" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }

  tags = {
    Tier = "Private"
  }
}

data "aws_subnets" "all" {}
`, rInt))
}

func testAccDataSourceAwsSubnetsConfig(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[1]d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnets-data-source"
  }
}

resource "aws_subnet" "test_public_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnets-data-source-public-a"
    Tier = "Public"
  }
}

resource "aws_subnet" "test_private_a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.125.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnets-data-source-private-a"
    Tier = "Private"
  }
}

resource "aws_subnet" "test_private_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.126.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-subnets-data-source-private-b"
    Tier = "Private"
  }
}
`, rInt))
}

func testAccDataSourceAwsSubnets_filter(rInt int) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.%[1]d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnets-data-source"
  }
}

resource "aws_subnet" "test_a_one" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test_a_two" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test_b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%[1]d.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

data "aws_subnets" "test_vpc_id" {
  filter {
    name   = "availabilityZone"
    values = [aws_subnet.test_a_one.availability_zone]
  }

  filter {
    name   = "vpc-id"
    values = [aws_subnet.test_a_two.vpc_id]
  }
}

data "aws_subnets" "test_no_vpc_id" {
  filter {
    name   = "availabilityZone"
    values = [aws_subnet.test_b.availability_zone]
  }
}
`, rInt))
}
