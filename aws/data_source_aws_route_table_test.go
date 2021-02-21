package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsRouteTable_basic(t *testing.T) {
	rtResourceName := "aws_route_table.test"
	snResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"
	igwResourceName := "aws_internet_gateway.test"
	datasource1Name := "data.aws_route_table.by_tag"
	datasource2Name := "data.aws_route_table.by_filter"
	datasource3Name := "data.aws_route_table.by_subnet"
	datasource4Name := "data.aws_route_table.by_id"
	datasource5Name := "data.aws_route_table.by_gateway"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					// By tags.
					testAccMatchResourceAttrRegionalARN(datasource1Name, "arn", "ec2", regexp.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource1Name, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(datasource1Name, "subnet_id"),
					resource.TestCheckNoResourceAttr(datasource1Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource1Name, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(datasource1Name, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(datasource1Name, "associations", "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource1Name, "tags.Name", rName),
					// By filter.
					testAccMatchResourceAttrRegionalARN(datasource2Name, "arn", "ec2", regexp.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource2Name, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(datasource2Name, "subnet_id"),
					resource.TestCheckNoResourceAttr(datasource2Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource2Name, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(datasource2Name, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(datasource2Name, "associations", "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource2Name, "tags.Name", rName),
					// By subnet ID.
					testAccMatchResourceAttrRegionalARN(datasource3Name, "arn", "ec2", regexp.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource3Name, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "subnet_id", snResourceName, "id"),
					resource.TestCheckNoResourceAttr(datasource3Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource3Name, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(datasource3Name, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(datasource3Name, "associations", "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource3Name, "tags.Name", rName),
					// By route table ID.
					testAccMatchResourceAttrRegionalARN(datasource4Name, "arn", "ec2", regexp.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource4Name, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource4Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource4Name, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasource4Name, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(datasource4Name, "subnet_id"),
					resource.TestCheckNoResourceAttr(datasource4Name, "gateway_id"),
					resource.TestCheckResourceAttr(datasource4Name, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(datasource4Name, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(datasource4Name, "associations", "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource4Name, "tags.Name", rName),
					// By gateway ID.
					testAccMatchResourceAttrRegionalARN(datasource5Name, "arn", "ec2", regexp.MustCompile(`route-table/.+$`)),
					resource.TestCheckResourceAttrPair(datasource5Name, "id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource5Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource5Name, "owner_id", rtResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasource5Name, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckNoResourceAttr(datasource5Name, "subnet_id"),
					resource.TestCheckResourceAttrPair(datasource5Name, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource5Name, "associations.#", "2"),
					testAccCheckListHasSomeElementAttrPair(datasource5Name, "associations", "subnet_id", snResourceName, "id"),
					testAccCheckListHasSomeElementAttrPair(datasource5Name, "associations", "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(datasource5Name, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRouteTable_main(t *testing.T) {
	datasourceName := "data.aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteTableConfigMain(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "vpc_id"),
					resource.TestCheckResourceAttr(datasourceName, "associations.0.main", "true"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRouteTableConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
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

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_tag" {
  tags = {
    Name = aws_route_table.test.tags["Name"]
  }

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_subnet" {
  subnet_id = aws_subnet.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_gateway" {
  gateway_id = aws_internet_gateway.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}

data "aws_route_table" "by_id" {
  route_table_id = aws_route_table.test.id

  depends_on = [aws_route_table_association.a, aws_route_table_association.b]
}
`, rName)
}

func testAccDataSourceAwsRouteTableConfigMain(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_route_table" "test" {
  filter {
    name   = "association.main"
    values = ["true"]
  }

  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }
}
`, rName)
}
