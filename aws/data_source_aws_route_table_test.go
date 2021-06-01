package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsRouteTable_basic(t *testing.T) {
	rtResourceName := "aws_route_table.test"
	snResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	gwResourceName := "aws_internet_gateway.test"
	ds1ResourceName := "data.aws_route_table.by_tag"
	ds2ResourceName := "data.aws_route_table.by_filter"
	ds3ResourceName := "data.aws_route_table.by_subnet"
	ds4ResourceName := "data.aws_route_table.by_id"
	ds5ResourceName := "data.aws_route_table.by_gateway"
	tagValue := "terraform-testacc-routetable-data-source"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteTableGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						ds1ResourceName, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds1ResourceName, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds1ResourceName, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(
						ds1ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(
						ds1ResourceName, "subnet_id"),
					resource.TestCheckNoResourceAttr(
						ds1ResourceName, "gateway_id"),
					resource.TestCheckResourceAttr(
						ds1ResourceName, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(
						ds1ResourceName, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(
						ds1ResourceName, "associations", "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds1ResourceName, "tags.Name", tagValue),

					resource.TestCheckResourceAttrPair(
						ds2ResourceName, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds2ResourceName, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds2ResourceName, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(
						ds2ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(
						ds2ResourceName, "subnet_id"),
					resource.TestCheckNoResourceAttr(
						ds2ResourceName, "gateway_id"),
					resource.TestCheckResourceAttr(
						ds2ResourceName, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(
						ds2ResourceName, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(
						ds2ResourceName, "associations", "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds2ResourceName, "tags.Name", tagValue),

					resource.TestCheckResourceAttrPair(
						ds3ResourceName, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds3ResourceName, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds3ResourceName, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(
						ds3ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds3ResourceName, "subnet_id", snResourceName, "id"),
					resource.TestCheckNoResourceAttr(
						ds3ResourceName, "gateway_id"),
					resource.TestCheckResourceAttr(
						ds3ResourceName, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(
						ds3ResourceName, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(
						ds3ResourceName, "associations", "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds3ResourceName, "tags.Name", tagValue),

					resource.TestCheckResourceAttrPair(
						ds4ResourceName, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds4ResourceName, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds4ResourceName, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(
						ds4ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(
						ds4ResourceName, "subnet_id"),
					resource.TestCheckNoResourceAttr(
						ds4ResourceName, "gateway_id"),
					resource.TestCheckResourceAttr(
						ds4ResourceName, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(
						ds4ResourceName, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(
						ds4ResourceName, "associations", "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds4ResourceName, "tags.Name", tagValue),

					resource.TestCheckResourceAttrPair(
						ds5ResourceName, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds5ResourceName, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						ds5ResourceName, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(
						ds5ResourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(
						ds5ResourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(
						ds5ResourceName, "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds5ResourceName, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(
						ds5ResourceName, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(
						ds5ResourceName, "associations", "gateway_id", gwResourceName, "id"),
					resource.TestCheckResourceAttr(
						ds5ResourceName, "tags.Name", tagValue),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSourceAwsRouteTable_main(t *testing.T) {
	dsResourceName := "data.aws_route_table.by_filter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteTableMainRoute,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						dsResourceName, "id"),
					resource.TestCheckResourceAttrSet(
						dsResourceName, "vpc_id"),
					resource.TestCheckResourceAttr(
						dsResourceName, "associations.0.main", "true"),
				),
			},
		},
	})
}

const testAccDataSourceAwsRouteTableGroupConfig = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-data-source"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-route-table-data-source"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-routetable-data-source"
  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-routetable-data-source"
  }
}

resource "aws_route_table_association" "b" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id
}

data "aws_route_table" "by_filter" {
  filter {
    name   = "association.route-table-association-id"
    values = [aws_route_table_association.a.id]
  }

  depends_on = [
    "aws_route_table_association.a",
    "aws_route_table_association.b"
  ]
}

data "aws_route_table" "by_tag" {
  tags = {
    Name = aws_route_table.test.tags["Name"]
  }

  depends_on = [
    "aws_route_table_association.a",
    "aws_route_table_association.b"
  ]
}

data "aws_route_table" "by_subnet" {
  subnet_id = aws_subnet.test.id
  depends_on = [
    "aws_route_table_association.a",
    "aws_route_table_association.b"
  ]
}

data "aws_route_table" "by_gateway" {
  gateway_id = aws_internet_gateway.test.id
  depends_on = [
    "aws_route_table_association.a",
    "aws_route_table_association.b"
  ]
}

data "aws_route_table" "by_id" {
  route_table_id = aws_route_table.test.id
  depends_on = [
    "aws_route_table_association.a",
    "aws_route_table_association.b"
  ]
}
`

const testAccDataSourceAwsRouteTableMainRoute = `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-data-source-main-route"
  }
}

data "aws_route_table" "by_filter" {
  filter {
    name   = "association.main"
    values = ["true"]
  }

  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}
`
