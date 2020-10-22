package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRouteDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRouteCheck("data.aws_route.by_destination_cidr_block"),
					testAccDataSourceAwsRouteCheck("data.aws_route.by_instance_id"),
					testAccDataSourceAwsRouteCheck("data.aws_route.by_peering_connection_id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRouteDataSource_TransitGatewayID(t *testing.T) {
	var route ec2.Route
	dataSourceName := "data.aws_route.test"
	resourceName := "aws_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigTransitGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "destination_cidr_block", dataSourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_LocalGatewayID(t *testing.T) {
	var route ec2.Route
	dataSourceName := "data.aws_route.by_local_gateway_id"
	resourceName := "aws_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigLocalGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "destination_cidr_block", dataSourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", dataSourceName, "local_gateway_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRouteCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		r, ok := s.RootModule().Resources["aws_route.test"]
		if !ok {
			return fmt.Errorf("can't find aws_route.test in state")
		}
		rts, ok := s.RootModule().Resources["aws_route_table.test"]
		if !ok {
			return fmt.Errorf("can't find aws_route_table.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["route_table_id"] != r.Primary.Attributes["route_table_id"] {
			return fmt.Errorf(
				"route_table_id is %s; want %s",
				attr["route_table_id"],
				r.Primary.Attributes["route_table_id"],
			)
		}

		if attr["route_table_id"] != rts.Primary.Attributes["id"] {
			return fmt.Errorf(
				"route_table_id is %s; want %s",
				attr["route_table_id"],
				rts.Primary.Attributes["id"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsRouteGroupConfig() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-data-source"
  }
}

resource "aws_vpc" "dest" {
  cidr_block = "172.17.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  peer_vpc_id = aws_vpc.dest.id
  vpc_id      = aws_vpc.test.id
  auto_accept = true
}

resource "aws_subnet" "test" {
  cidr_block        = "172.16.0.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

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

resource "aws_route" "pcx" {
  route_table_id            = aws_route_table.test.id
  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  destination_cidr_block    = "10.0.2.0/24"
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.0.1.0/24"
  instance_id            = aws_instance.web.id

  timeouts {
    create = "5m"
  }
}

data "aws_route" "by_peering_connection_id" {
  route_table_id            = aws_route_table.test.id
  vpc_peering_connection_id = aws_route.pcx.vpc_peering_connection_id
}

data "aws_route" "by_destination_cidr_block" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.0.1.0/24"
  depends_on             = [aws_route.test]
}

data "aws_route" "by_instance_id" {
  route_table_id = aws_route_table.test.id
  instance_id    = aws_instance.web.id
  depends_on     = [aws_route.test]
}
`)
}

func testAccAWSRouteDataSourceConfigTransitGatewayID() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `
# IncorrectState: Transit Gateway is not available in some availability zones

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-datasource-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-route-datasource-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  route_table_id         = aws_vpc.test.default_route_table_id
  transit_gateway_id     = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}

data "aws_route" "test" {
  route_table_id     = aws_route.test.route_table_id
  transit_gateway_id = aws_route.test.transit_gateway_id
}
`
}

func testAccAWSRouteDataSourceConfigLocalGatewayID() string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "all" {}
data "aws_ec2_local_gateway" "first" {
  id = tolist(data.aws_ec2_local_gateways.all.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "all" {}
data "aws_ec2_local_gateway_route_table" "first" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.all.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "172.16.1.0/24"
  local_gateway_id       = data.aws_ec2_local_gateway.first.id
  depends_on             = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

data "aws_route" "by_local_gateway_id" {
  route_table_id   = aws_route_table.test.id
  local_gateway_id = data.aws_ec2_local_gateway.first.id
  depends_on       = [aws_route.test]
}
`)
}
