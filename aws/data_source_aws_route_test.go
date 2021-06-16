package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRouteDataSource_basic(t *testing.T) {
	instanceRouteResourceName := "aws_route.instance"
	pcxRouteResourceName := "aws_route.vpc_peering_connection"
	rtResourceName := "aws_route_table.test"
	instanceResourceName := "aws_instance.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	datasource1Name := "data.aws_route.by_destination_cidr_block"
	datasource2Name := "data.aws_route.by_instance_id"
	datasource3Name := "data.aws_route.by_peering_connection_id"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					// By destination CIDR.
					resource.TestCheckResourceAttrPair(datasource1Name, "destination_cidr_block", instanceRouteResourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource1Name, "route_table_id", rtResourceName, "id"),

					// By instance ID.
					resource.TestCheckResourceAttrPair(datasource2Name, "destination_cidr_block", instanceRouteResourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource2Name, "instance_id", instanceResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "route_table_id", rtResourceName, "id"),

					// By VPC peering connection ID.
					resource.TestCheckResourceAttrPair(datasource3Name, "destination_cidr_block", pcxRouteResourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource3Name, "route_table_id", rtResourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_TransitGatewayID(t *testing.T) {
	dataSourceName := "data.aws_route.test"
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigIpv4TransitGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "destination_cidr_block", dataSourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_IPv6DestinationCidr(t *testing.T) {
	dataSourceName := "data.aws_route.test"
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigIpv6EgressOnlyInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "destination_ipv6_cidr_block", dataSourceName, "destination_ipv6_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_LocalGatewayID(t *testing.T) {
	dataSourceName := "data.aws_route.by_local_gateway_id"
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigIpv4LocalGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "destination_cidr_block", dataSourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", dataSourceName, "local_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_CarrierGatewayID(t *testing.T) {
	dataSourceName := "data.aws_route.test"
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWavelengthZoneAvailable(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigIpv4CarrierGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "destination_cidr_block", dataSourceName, "destination_cidr_block"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "carrier_gateway_id", dataSourceName, "carrier_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_DestinationPrefixListId(t *testing.T) {
	dataSourceName := "data.aws_route.test"
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigPrefixListNatGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", dataSourceName, "destination_prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", dataSourceName, "nat_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "route_table_id", dataSourceName, "route_table_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteDataSource_GatewayVpcEndpoint(t *testing.T) {
	var routeTable ec2.RouteTable
	var vpce ec2.VpcEndpoint
	rtResourceName := "aws_route_table.test"
	vpceResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteDataSourceConfigGatewayVpcEndpointNoDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(rtResourceName, &routeTable),
					testAccCheckVpcEndpointExists(vpceResourceName, &vpce),
					testAccCheckAWSRouteTableWaitForVpcEndpointRoute(&routeTable, &vpce),
				),
			},
			{
				Config:      testAccAWSRouteDataSourceConfigGatewayVpcEndpointWithDataSource(rName),
				ExpectError: regexp.MustCompile(`No routes matching supplied arguments found in Route Table`),
			},
		},
	})
}

func testAccDataSourceAwsRouteConfigBasic(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "172.17.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  peer_vpc_id = aws_vpc.target.id
  vpc_id      = aws_vpc.test.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "172.16.0.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

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

resource "aws_route" "vpc_peering_connection" {
  route_table_id            = aws_route_table.test.id
  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  destination_cidr_block    = "10.0.2.0/24"
}

resource "aws_route_table_association" "a" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "instance" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.0.1.0/24"
  instance_id            = aws_instance.test.id
}

data "aws_route" "by_peering_connection_id" {
  route_table_id            = aws_route_table.test.id
  vpc_peering_connection_id = aws_route.vpc_peering_connection.vpc_peering_connection_id
}

data "aws_route" "by_destination_cidr_block" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = aws_route.instance.destination_cidr_block
}

data "aws_route" "by_instance_id" {
  route_table_id = aws_route_table.test.id
  instance_id    = aws_route.instance.instance_id
}
`, rName))
}

func testAccAWSRouteDataSourceConfigIpv4TransitGateway(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
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
`, rName))
}

func testAccAWSRouteDataSourceConfigIpv6EgressOnlyInternetGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = "::/0"
  egress_only_gateway_id      = aws_egress_only_internet_gateway.test.id
}

data "aws_route" "test" {
  route_table_id              = aws_route.test.route_table_id
  destination_ipv6_cidr_block = aws_route.test.destination_ipv6_cidr_block
}
`, rName)
}

func testAccAWSRouteDataSourceConfigIpv4LocalGateway(rName string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

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
`, rName)
}

func testAccAWSRouteDataSourceConfigIpv4CarrierGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  route_table_id         = aws_route_table.test.id
  carrier_gateway_id     = aws_ec2_carrier_gateway.test.id
}

data "aws_route" "test" {
  route_table_id     = aws_route.test.route_table_id
  carrier_gateway_id = aws_route.test.carrier_gateway_id
}
`, rName)
}

func testAccAWSRouteDataSourceConfigPrefixListNatGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  nat_gateway_id             = aws_nat_gateway.test.id
}

data "aws_route" "test" {
  route_table_id             = aws_route.test.route_table_id
  destination_prefix_list_id = aws_route.test.destination_prefix_list_id
  nat_gateway_id             = aws_route.test.nat_gateway_id
}
`, rName)
}

func testAccAWSRouteDataSourceConfigGatewayVpcEndpointNoDataSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteDataSourceConfigGatewayVpcEndpointWithDataSource(rName string) string {
	return composeConfig(testAccAWSRouteDataSourceConfigGatewayVpcEndpointNoDataSource(rName), `
data "aws_prefix_list" "test" {
  name = aws_vpc_endpoint.test.service_name
}

data "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = data.aws_prefix_list.test.id
}
  `)
}
