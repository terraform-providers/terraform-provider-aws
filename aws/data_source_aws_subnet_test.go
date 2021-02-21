package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsSubnet_basic(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	cidr := fmt.Sprintf("172.%d.123.0/24", rInt)
	tag := "tf-acc-subnet-data-source"

	snResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_subnet.by_id"
	ds2ResourceName := "data.aws_subnet.by_cidr"
	ds3ResourceName := "data.aws_subnet.by_tag"
	ds4ResourceName := "data.aws_subnet.by_vpc"
	ds5ResourceName := "data.aws_subnet.by_filter"
	ds6ResourceName := "data.aws_subnet.by_az_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds1ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds1ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds1ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds2ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds2ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds2ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(ds3ResourceName, "available_ip_address_count"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds3ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds3ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds4ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds4ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds4ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds4ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds5ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds5ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds5ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds5ResourceName, "outpost_arn", snResourceName, "outpost_arn"),

					resource.TestCheckResourceAttrPair(ds6ResourceName, "id", snResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "owner_id", snResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "availability_zone", snResourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "availability_zone_id", snResourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(ds6ResourceName, "cidr_block", cidr),
					resource.TestCheckResourceAttr(ds6ResourceName, "tags.Name", tag),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "arn", snResourceName, "arn"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "customer_owned_ipv4_pool", snResourceName, "customer_owned_ipv4_pool"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "map_customer_owned_ip_on_launch", snResourceName, "map_customer_owned_ip_on_launch"),
					resource.TestCheckResourceAttrPair(ds6ResourceName, "outpost_arn", snResourceName, "outpost_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSubnet_ipv6ByIpv6Filter(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetConfigIpv6(rInt),
			},
			{
				Config: testAccDataSourceAwsSubnetConfigIpv6WithDataSourceFilter(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSubnet_ipv6ByIpv6CidrBlock(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSubnetConfigIpv6(rInt),
			},
			{
				Config: testAccDataSourceAwsSubnetConfigIpv6WithDataSourceIpv6CidrBlock(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_subnet.by_ipv6_cidr", "ipv6_cidr_block_association_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSubnetConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "172.%d.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-data-source"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-subnet-data-source"
  }
}

data "aws_subnet" "by_id" {
  id = aws_subnet.test.id
}

data "aws_subnet" "by_cidr" {
  vpc_id     = aws_subnet.test.vpc_id
  cidr_block = aws_subnet.test.cidr_block
}

data "aws_subnet" "by_tag" {
  vpc_id = aws_subnet.test.vpc_id

  tags = {
    Name = aws_subnet.test.tags["Name"]
  }
}

data "aws_subnet" "by_vpc" {
  vpc_id = aws_subnet.test.vpc_id
}

data "aws_subnet" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_subnet.test.vpc_id]
  }
}

data "aws_subnet" "by_az_id" {
  vpc_id               = aws_subnet.test.vpc_id
  availability_zone_id = aws_subnet.test.availability_zone_id
}
`, rInt, rInt)
}

func testAccDataSourceAwsSubnetConfigIpv6(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-data-source-ipv6"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-subnet-data-source-ipv6"
  }
}
`, rInt, rInt)
}

func testAccDataSourceAwsSubnetConfigIpv6WithDataSourceFilter(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-data-source-ipv6-with-ds-filter"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-subnet-data-source-ipv6-with-ds-filter"
  }
}

data "aws_subnet" "by_ipv6_cidr" {
  filter {
    name   = "ipv6-cidr-block-association.ipv6-cidr-block"
    values = [aws_subnet.test.ipv6_cidr_block]
  }
}
`, rInt, rInt)
}

func testAccDataSourceAwsSubnetConfigIpv6WithDataSourceIpv6CidrBlock(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block                       = "172.%d.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-subnet-data-source-ipv6-cidr-block"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "172.%d.123.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = "tf-acc-subnet-data-source-ipv6-cidr-block"
  }
}

data "aws_subnet" "by_ipv6_cidr" {
  ipv6_cidr_block = aws_subnet.test.ipv6_cidr_block
}
`, rInt, rInt)
}
