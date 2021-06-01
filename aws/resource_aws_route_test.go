package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRoute_basic(t *testing.T) {
	var route ec2.Route

	//aws creates a default route
	testCheck := func(s *terraform.State) error {
		if *route.DestinationCidrBlock != "10.3.0.0/16" {
			return fmt.Errorf("Destination Cidr (Expected=%s, Actual=%s)\n", "10.3.0.0/16", *route.DestinationCidrBlock)
		}

		name := "aws_internet_gateway.foo"
		gwres, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s\n", name)
		}

		if *route.GatewayId != gwres.Primary.ID {
			return fmt.Errorf("Internet Gateway Id (Expected=%s, Actual=%s)\n", gwres.Primary.ID, *route.GatewayId)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testCheck,
				),
			},
			{
				ResourceName:      "aws_route.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_disappears(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute(), "aws_route.bar"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute_ipv6Support(t *testing.T) {
	var route ec2.Route

	//aws creates a default route
	testCheck := func(s *terraform.State) error {
		name := "aws_egress_only_internet_gateway.foo"
		gwres, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s\n", name)
		}

		if *route.EgressOnlyInternetGatewayId != gwres.Primary.ID {
			return fmt.Errorf("Egress Only Internet Gateway Id (Expected=%s, Actual=%s)\n", gwres.Primary.ID, *route.EgressOnlyInternetGatewayId)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testCheck,
					resource.TestCheckResourceAttr("aws_route.bar", "destination_ipv6_cidr_block", "::/0"),
				),
			},
			{
				ResourceName:      "aws_route.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.bar"),
				ImportStateVerify: true,
			},
			{
				Config:   testAccAWSRouteConfigIpv6Expanded(),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSRoute_ipv6ToInternetGateway(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6InternetGateway(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.igw", &route),
				),
			},
			{
				ResourceName:      "aws_route.igw",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.igw"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_ipv6ToInstance(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6Instance(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.internal-default-route-ipv6", &route),
				),
			},
			{
				ResourceName:      "aws_route.internal-default-route-ipv6",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.internal-default-route-ipv6"),
				ImportStateVerify: true,
			},
			{
				Config:   testAccAWSRouteConfigIpv6InstanceExpanded(),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSRoute_ipv6ToNetworkInterface(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6NetworkInterface(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.internal-default-route-ipv6", &route),
				),
			},
			{
				ResourceName:      "aws_route.internal-default-route-ipv6",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.internal-default-route-ipv6"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_ipv6ToPeeringConnection(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6PeeringConnection(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.pc", &route),
				),
			},
			{
				ResourceName:      "aws_route.pc",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.pc"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_changeRouteTable(t *testing.T) {
	var before ec2.Route
	var after ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &before),
				),
			},
			{
				Config: testAccAWSRouteNewRouteTable(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &after),
				),
			},
			{
				ResourceName:      "aws_route.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_changeCidr(t *testing.T) {
	var route ec2.Route
	var routeTable ec2.RouteTable

	//aws creates a default route
	testCheck := func(s *terraform.State) error {
		if *route.DestinationCidrBlock != "10.3.0.0/16" {
			return fmt.Errorf("Destination Cidr (Expected=%s, Actual=%s)\n", "10.3.0.0/16", *route.DestinationCidrBlock)
		}

		name := "aws_internet_gateway.foo"
		gwres, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s\n", name)
		}

		if *route.GatewayId != gwres.Primary.ID {
			return fmt.Errorf("Internet Gateway Id (Expected=%s, Actual=%s)\n", gwres.Primary.ID, *route.GatewayId)
		}

		return nil
	}

	testCheckChange := func(s *terraform.State) error {
		if *route.DestinationCidrBlock != "10.2.0.0/16" {
			return fmt.Errorf("Destination Cidr (Expected=%s, Actual=%s)\n", "10.2.0.0/16", *route.DestinationCidrBlock)
		}

		name := "aws_internet_gateway.foo"
		gwres, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s\n", name)
		}

		if *route.GatewayId != gwres.Primary.ID {
			return fmt.Errorf("Internet Gateway Id (Expected=%s, Actual=%s)\n", gwres.Primary.ID, *route.GatewayId)
		}

		if rtlen := len(routeTable.Routes); rtlen != 2 {
			return fmt.Errorf("Route Table has too many routes (Expected=%d, Actual=%d)\n", rtlen, 2)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testCheck,
				),
			},
			{
				Config: testAccAWSRouteBasicConfigChangeCidr(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testAccCheckRouteTableExists("aws_route_table.foo", &routeTable),
					testCheckChange,
				),
			},
			{
				ResourceName:      "aws_route.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_noopdiff(t *testing.T) {
	var route ec2.Route
	var routeTable ec2.RouteTable

	testCheck := func(s *terraform.State) error {
		return nil
	}

	testCheckChange := func(s *terraform.State) error {
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteNoopChange(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.test", &route),
					testCheck,
				),
			},
			{
				Config: testAccAWSRouteNoopChange(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.test", &route),
					testAccCheckRouteTableExists("aws_route_table.test", &routeTable),
					testCheckChange,
				),
			},
			{
				ResourceName:      "aws_route.test",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.test"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_doesNotCrashWithVPCEndpoint(t *testing.T) {
	var route ec2.Route

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteWithVPCEndpoint(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
				),
			},
			{
				ResourceName:      "aws_route.bar",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc("aws_route.bar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_TransitGatewayID_DestinationCidrBlock(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigTransitGatewayIDDestinatationCidrBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_LocalGatewayID(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteResourceConfigLocalGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_ConditionalCidrBlock(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigConditionalIpv4Ipv6(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigConditionalIpv4Ipv6(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", "::/0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_VpcEndpointId(t *testing.T) {
	var route ec2.Route
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteResourceConfigVpcEndpointId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_id", vpcEndpointResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSRouteExists(n string, res *ec2.Route) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s\n", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		r, err := resourceAwsRouteFindRoute(
			conn,
			rs.Primary.Attributes["route_table_id"],
			rs.Primary.Attributes["destination_cidr_block"],
			rs.Primary.Attributes["destination_ipv6_cidr_block"],
		)

		if err != nil {
			return err
		}

		if r == nil {
			return fmt.Errorf("Route not found")
		}

		*res = *r

		return nil
	}
}

func testAccCheckAWSRouteDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		route, err := resourceAwsRouteFindRoute(
			conn,
			rs.Primary.Attributes["route_table_id"],
			rs.Primary.Attributes["destination_cidr_block"],
			rs.Primary.Attributes["destination_ipv6_cidr_block"],
		)

		if route == nil && err == nil {
			return nil
		}
	}

	return nil
}

func testAccAWSRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		destination := rs.Primary.Attributes["destination_cidr_block"]
		if v, ok := rs.Primary.Attributes["destination_ipv6_cidr_block"]; ok && v != "" {
			destination = v
		}

		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["route_table_id"], destination), nil
	}
}

func testAccAWSRouteBasicConfig() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-basic"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-basic"
  }
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "bar" {
  route_table_id         = aws_route_table.foo.id
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = aws_internet_gateway.foo.id
}
`
}

func testAccAWSRouteConfigIpv6InternetGateway() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6-igw"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-ipv6-igw"
  }
}

resource "aws_route_table" "external" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "igw" {
  route_table_id              = aws_route_table.external.id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.foo.id
}
`
}

func testAccAWSRouteConfigIpv6NetworkInterface() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.router-network.availability_zone", "t2.small", "t3.small"),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "examplevpc" {
  cidr_block                       = "10.100.0.0/16"
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6-network-interface"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_internet_gateway" "internet" {
  vpc_id = aws_vpc.examplevpc.id

  tags = {
    Name = "terraform-testacc-route-ipv6-network-interface"
  }
}

resource "aws_route" "igw" {
  route_table_id         = aws_vpc.examplevpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet.id
}

resource "aws_route" "igw-ipv6" {
  route_table_id              = aws_vpc.examplevpc.main_route_table_id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.internet.id
}

resource "aws_subnet" "router-network" {
  cidr_block                      = "10.100.1.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = true
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-network-interface-router"
  }
}

resource "aws_subnet" "client-network" {
  cidr_block                      = "10.100.10.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 2)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = false
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-network-interface-client"
  }
}

resource "aws_route_table" "client-routes" {
  vpc_id = aws_vpc.examplevpc.id
}

resource "aws_route_table_association" "client-routes" {
  route_table_id = aws_route_table.client-routes.id
  subnet_id      = aws_subnet.client-network.id
}

resource "aws_instance" "test-router" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.router-network.id
}

resource "aws_network_interface" "router-internal" {
  subnet_id         = aws_subnet.client-network.id
  source_dest_check = false
}

resource "aws_network_interface_attachment" "router-internal" {
  device_index         = 1
  instance_id          = aws_instance.test-router.id
  network_interface_id = aws_network_interface.router-internal.id
}

resource "aws_route" "internal-default-route" {
  route_table_id         = aws_route_table.client-routes.id
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id   = aws_network_interface.router-internal.id
}

resource "aws_route" "internal-default-route-ipv6" {
  route_table_id              = aws_route_table.client-routes.id
  destination_ipv6_cidr_block = "::/0"
  network_interface_id        = aws_network_interface.router-internal.id
}
`)
}

func testAccAWSRouteConfigIpv6Instance() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.router-network.availability_zone", "t2.small", "t3.small"),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "examplevpc" {
  cidr_block                       = "10.100.0.0/16"
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_internet_gateway" "internet" {
  vpc_id = aws_vpc.examplevpc.id

  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

resource "aws_route" "igw" {
  route_table_id         = aws_vpc.examplevpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet.id
}

resource "aws_route" "igw-ipv6" {
  route_table_id              = aws_vpc.examplevpc.main_route_table_id
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.internet.id
}

resource "aws_subnet" "router-network" {
  cidr_block                      = "10.100.1.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = true
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-instance-router"
  }
}

resource "aws_subnet" "client-network" {
  cidr_block                      = "10.100.10.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 2)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = false
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-instance-client"
  }
}

resource "aws_route_table" "client-routes" {
  vpc_id = aws_vpc.examplevpc.id
}

resource "aws_route_table_association" "client-routes" {
  route_table_id = aws_route_table.client-routes.id
  subnet_id      = aws_subnet.client-network.id
}

resource "aws_instance" "test-router" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.router-network.id
}

resource "aws_route" "internal-default-route" {
  route_table_id         = aws_route_table.client-routes.id
  destination_cidr_block = "0.0.0.0/0"
  instance_id            = aws_instance.test-router.id
}

resource "aws_route" "internal-default-route-ipv6" {
  route_table_id              = aws_route_table.client-routes.id
  destination_ipv6_cidr_block = "::/0"
  instance_id                 = aws_instance.test-router.id
}
`)
}

func testAccAWSRouteConfigIpv6InstanceExpanded() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.router-network.availability_zone", "t2.small", "t3.small"),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
resource "aws_vpc" "examplevpc" {
  cidr_block                       = "10.100.0.0/16"
  enable_dns_hostnames             = true
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_internet_gateway" "internet" {
  vpc_id = aws_vpc.examplevpc.id

  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

resource "aws_route" "igw" {
  route_table_id         = aws_vpc.examplevpc.main_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet.id
}

resource "aws_route" "igw-ipv6" {
  route_table_id              = aws_vpc.examplevpc.main_route_table_id
  destination_ipv6_cidr_block = "::0/0"
  gateway_id                  = aws_internet_gateway.internet.id
}

resource "aws_subnet" "router-network" {
  cidr_block                      = "10.100.1.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = true
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-instance-router"
  }
}

resource "aws_subnet" "client-network" {
  cidr_block                      = "10.100.10.0/24"
  vpc_id                          = aws_vpc.examplevpc.id
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 2)
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch         = false
  availability_zone               = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-route-ipv6-instance-client"
  }
}

resource "aws_route_table" "client-routes" {
  vpc_id = aws_vpc.examplevpc.id
}

resource "aws_route_table_association" "client-routes" {
  route_table_id = aws_route_table.client-routes.id
  subnet_id      = aws_subnet.client-network.id
}

resource "aws_instance" "test-router" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.router-network.id
}

resource "aws_route" "internal-default-route" {
  route_table_id         = aws_route_table.client-routes.id
  destination_cidr_block = "0.0.0.0/0"
  instance_id            = aws_instance.test-router.id
}

resource "aws_route" "internal-default-route-ipv6" {
  route_table_id              = aws_route_table.client-routes.id
  destination_ipv6_cidr_block = "::0/0"
  instance_id                 = aws_instance.test-router.id
}
`)
}

func testAccAWSRouteConfigIpv6PeeringConnection() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6-peering-connection"
  }
}

resource "aws_vpc" "bar" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_peering_connection" "foo" {
  vpc_id      = aws_vpc.foo.id
  peer_vpc_id = aws_vpc.bar.id
  auto_accept = true
}

resource "aws_route_table" "peering" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "pc" {
  route_table_id              = aws_route_table.peering.id
  destination_ipv6_cidr_block = aws_vpc.bar.ipv6_cidr_block
  vpc_peering_connection_id   = aws_vpc_peering_connection.foo.id
}
`
}

func testAccAWSRouteConfigIpv6() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "bar" {
  route_table_id              = aws_route_table.foo.id
  destination_ipv6_cidr_block = "::/0"
  egress_only_gateway_id      = aws_egress_only_internet_gateway.foo.id
}
`
}

func testAccAWSRouteConfigIpv6Expanded() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-ipv6"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "bar" {
  route_table_id              = aws_route_table.foo.id
  destination_ipv6_cidr_block = "::0/0"
  egress_only_gateway_id      = aws_egress_only_internet_gateway.foo.id
}
`
}

func testAccAWSRouteBasicConfigChangeCidr() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-change-cidr"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-change-cidr"
  }
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "bar" {
  route_table_id         = aws_route_table.foo.id
  destination_cidr_block = "10.2.0.0/16"
  gateway_id             = aws_internet_gateway.foo.id
}
`
}

func testAccAWSRouteNoopChange() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t2.nano", "t3.nano"),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"

  tags = {
    Name = "terraform-testacc-route-noop-change"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.10.10.0/24"

  tags = {
    Name = "tf-acc-route-noop-change"
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "0.0.0.0/0"
  instance_id            = aws_instance.nat.id
}

resource "aws_instance" "nat" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
}
`)
}

func testAccAWSRouteWithVPCEndpoint() string {
	return `
data "aws_region" "current" {}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-with-vpc-endpoint"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-with-vpc-endpoint"
  }
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_route" "bar" {
  route_table_id         = aws_route_table.foo.id
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = aws_internet_gateway.foo.id

  # Forcing endpoint to create before route - without this the crash is a race.
  depends_on = [aws_vpc_endpoint.baz]
}

resource "aws_vpc_endpoint" "baz" {
  vpc_id          = aws_vpc.foo.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.foo.id]
}
`
}

func testAccAWSRouteNewRouteTable() string {
	return `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-basic"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-route-new-route-table"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-basic"
  }
}

resource "aws_internet_gateway" "bar" {
  vpc_id = aws_vpc.bar.id

  tags = {
    Name = "terraform-testacc-route-new-route-table"
  }
}

resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "terraform-testacc-route-basic"
  }
}

resource "aws_route_table" "bar" {
  vpc_id = aws_vpc.bar.id

  tags = {
    Name = "terraform-testacc-route-new-route-table"
  }
}

resource "aws_route" "bar" {
  route_table_id         = aws_route_table.bar.id
  destination_cidr_block = "10.4.0.0/16"
  gateway_id             = aws_internet_gateway.bar.id
}
`
}

func testAccAWSRouteConfigTransitGatewayIDDestinatationCidrBlock() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
# IncorrectState: Transit Gateway is not available in availability zone usw2-az4	

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-route-transit-gateway-id"
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
`)
}

func testAccAWSRouteConfigConditionalIpv4Ipv6(rName string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

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

resource "aws_internet_gateway" "test" {
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

locals {
  ipv6             = %[2]t
  destination      = "0.0.0.0/0"
  destination_ipv6 = "::/0"
}

resource "aws_route" "test" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id

  destination_cidr_block      = local.ipv6 ? "" : local.destination
  destination_ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
}
`, rName, ipv6Route)
}

func testAccAWSRouteResourceConfigLocalGatewayID() string {
	return `
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
  vpc_id     = aws_vpc.test.id
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "172.16.1.0/24"
  local_gateway_id       = data.aws_ec2_local_gateway.first.id
}
`
}

func testAccAWSRouteResourceConfigVpcEndpointId(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_caller_identity.current.arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "172.16.1.0/24"
  vpc_endpoint_id        = aws_vpc_endpoint.test.id
}
`, rName))
}
