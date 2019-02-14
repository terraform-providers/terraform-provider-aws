package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
				Config: testAccAWSRouteBasicConfig,
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
				Config: testAccAWSRouteConfigIpv6,
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
				Config: testAccAWSRouteConfigIpv6InternetGateway,
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
				Config: testAccAWSRouteConfigIpv6Instance,
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
				Config: testAccAWSRouteConfigIpv6NetworkInterface,
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
				Config: testAccAWSRouteConfigIpv6PeeringConnection,
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
				Config: testAccAWSRouteBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &before),
				),
			},
			{
				Config: testAccAWSRouteNewRouteTable,
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
				Config: testAccAWSRouteBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.bar", &route),
					testCheck,
				),
			},
			{
				Config: testAccAWSRouteBasicConfigChangeCidr,
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
				Config: testAccAWSRouteNoopChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists("aws_route.test", &route),
					testCheck,
				),
			},
			{
				Config: testAccAWSRouteNoopChange,
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
				Config: testAccAWSRouteWithVPCEndpoint,
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
		if _, ok := rs.Primary.Attributes["destination_ipv6_cidr_block"]; ok {
			destination = rs.Primary.Attributes["destination_ipv6_cidr_block"]
		}

		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["route_table_id"], destination), nil
	}
}

var testAccAWSRouteBasicConfig = fmt.Sprint(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-basic"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-basic"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "bar" {
	route_table_id = "${aws_route_table.foo.id}"
	destination_cidr_block = "10.3.0.0/16"
	gateway_id = "${aws_internet_gateway.foo.id}"
}
`)

var testAccAWSRouteConfigIpv6InternetGateway = fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "terraform-testacc-route-ipv6-igw"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-ipv6-igw"
	}
}

resource "aws_route_table" "external" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "igw" {
  route_table_id = "${aws_route_table.external.id}"
  destination_ipv6_cidr_block = "::/0"
  gateway_id = "${aws_internet_gateway.foo.id}"
}

`)

var testAccAWSRouteConfigIpv6NetworkInterface = fmt.Sprintf(`
resource "aws_vpc" "examplevpc" {
  cidr_block = "10.100.0.0/16"
  enable_dns_hostnames = true
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "terraform-testacc-route-ipv6-network-interface"
  }
}

data "aws_availability_zones" "available" {}

resource "aws_internet_gateway" "internet" {
  vpc_id = "${aws_vpc.examplevpc.id}"

  tags = {
    Name = "terraform-testacc-route-ipv6-network-interface"
  }
}

resource "aws_route" "igw" {
  route_table_id = "${aws_vpc.examplevpc.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.internet.id}"
}

resource "aws_route" "igw-ipv6" {
  route_table_id = "${aws_vpc.examplevpc.main_route_table_id}"
  destination_ipv6_cidr_block = "::/0"
  gateway_id = "${aws_internet_gateway.internet.id}"
}

resource "aws_subnet" "router-network" {
  cidr_block = "10.100.1.0/24"
  vpc_id = "${aws_vpc.examplevpc.id}"
  ipv6_cidr_block = "${cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 1)}"
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch = true
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-route-ipv6-network-interface-router"
  }
}

resource "aws_subnet" "client-network" {
  cidr_block = "10.100.10.0/24"
  vpc_id = "${aws_vpc.examplevpc.id}"
  ipv6_cidr_block = "${cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 2)}"
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch = false
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-route-ipv6-network-interface-client"
  }
}

resource "aws_route_table" "client-routes" {
  vpc_id = "${aws_vpc.examplevpc.id}"
}

resource "aws_route_table_association" "client-routes" {
  route_table_id = "${aws_route_table.client-routes.id}"
  subnet_id = "${aws_subnet.client-network.id}"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
      name   = "name"
      values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }
  filter {
      name   = "virtualization-type"
      values = ["hvm"]
  }
  owners = ["099720109477"]
}

resource "aws_instance" "test-router" {
  ami = "${data.aws_ami.ubuntu.image_id}"
  instance_type = "t2.small"
  subnet_id = "${aws_subnet.router-network.id}"
}

resource "aws_network_interface" "router-internal" {
  subnet_id = "${aws_subnet.client-network.id}"
  source_dest_check = false
}

resource "aws_network_interface_attachment" "router-internal" {
  device_index = 1
  instance_id = "${aws_instance.test-router.id}"
  network_interface_id = "${aws_network_interface.router-internal.id}"
}

resource "aws_route" "internal-default-route" {
  route_table_id = "${aws_route_table.client-routes.id}"
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id = "${aws_network_interface.router-internal.id}"
}

resource "aws_route" "internal-default-route-ipv6" {
  route_table_id = "${aws_route_table.client-routes.id}"
  destination_ipv6_cidr_block = "::/0"
  network_interface_id = "${aws_network_interface.router-internal.id}"
}

`)

var testAccAWSRouteConfigIpv6Instance = fmt.Sprintf(`
resource "aws_vpc" "examplevpc" {
  cidr_block = "10.100.0.0/16"
  enable_dns_hostnames = true
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

data "aws_availability_zones" "available" {}

resource "aws_internet_gateway" "internet" {
  vpc_id = "${aws_vpc.examplevpc.id}"

  tags = {
    Name = "terraform-testacc-route-ipv6-instance"
  }
}

resource "aws_route" "igw" {
  route_table_id = "${aws_vpc.examplevpc.main_route_table_id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.internet.id}"
}

resource "aws_route" "igw-ipv6" {
  route_table_id = "${aws_vpc.examplevpc.main_route_table_id}"
  destination_ipv6_cidr_block = "::/0"
  gateway_id = "${aws_internet_gateway.internet.id}"
}

resource "aws_subnet" "router-network" {
  cidr_block = "10.100.1.0/24"
  vpc_id = "${aws_vpc.examplevpc.id}"
  ipv6_cidr_block = "${cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 1)}"
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch = true
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-route-ipv6-instance-router"
  }
}

resource "aws_subnet" "client-network" {
  cidr_block = "10.100.10.0/24"
  vpc_id = "${aws_vpc.examplevpc.id}"
  ipv6_cidr_block = "${cidrsubnet(aws_vpc.examplevpc.ipv6_cidr_block, 8, 2)}"
  assign_ipv6_address_on_creation = true
  map_public_ip_on_launch = false
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-route-ipv6-instance-client"
  }
}

resource "aws_route_table" "client-routes" {
  vpc_id = "${aws_vpc.examplevpc.id}"
}

resource "aws_route_table_association" "client-routes" {
  route_table_id = "${aws_route_table.client-routes.id}"
  subnet_id = "${aws_subnet.client-network.id}"
}

data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
      name   = "name"
      values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }
  filter {
      name   = "virtualization-type"
      values = ["hvm"]
  }
  owners = ["099720109477"]
}

resource "aws_instance" "test-router" {
  ami = "${data.aws_ami.ubuntu.image_id}"
  instance_type = "t2.small"
  subnet_id = "${aws_subnet.router-network.id}"
}

resource "aws_route" "internal-default-route" {
  route_table_id = "${aws_route_table.client-routes.id}"
  destination_cidr_block = "0.0.0.0/0"
  instance_id = "${aws_instance.test-router.id}"
}

resource "aws_route" "internal-default-route-ipv6" {
  route_table_id = "${aws_route_table.client-routes.id}"
  destination_ipv6_cidr_block = "::/0"
  instance_id = "${aws_instance.test-router.id}"
}

`)

var testAccAWSRouteConfigIpv6PeeringConnection = fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.0.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-route-ipv6-peering-connection"
	}
}

resource "aws_vpc" "bar" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
}

resource "aws_vpc_peering_connection" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
	peer_vpc_id = "${aws_vpc.bar.id}"
	auto_accept = true
}

resource "aws_route_table" "peering" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "pc" {
  route_table_id = "${aws_route_table.peering.id}"
  destination_ipv6_cidr_block = "${aws_vpc.bar.ipv6_cidr_block}"
  vpc_peering_connection_id = "${aws_vpc_peering_connection.foo.id}"
}

`)

var testAccAWSRouteConfigIpv6 = fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "terraform-testacc-route-ipv6"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "bar" {
	route_table_id = "${aws_route_table.foo.id}"
	destination_ipv6_cidr_block = "::/0"
	egress_only_gateway_id = "${aws_egress_only_internet_gateway.foo.id}"
}


`)

var testAccAWSRouteBasicConfigChangeCidr = fmt.Sprint(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-change-cidr"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-change-cidr"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "bar" {
	route_table_id = "${aws_route_table.foo.id}"
	destination_cidr_block = "10.2.0.0/16"
	gateway_id = "${aws_internet_gateway.foo.id}"
}
`)

var testAccAWSRouteNoopChange = fmt.Sprint(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-route-noop-change"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.10.10.0/24"
  tags = {
    Name = "tf-acc-route-noop-change"
  }
}

resource "aws_route" "test" {
  route_table_id = "${aws_route_table.test.id}"
  destination_cidr_block = "0.0.0.0/0"
  instance_id = "${aws_instance.nat.id}"
}

resource "aws_instance" "nat" {
  ami = "ami-9abea4fb"
  instance_type = "t2.nano"
  subnet_id = "${aws_subnet.test.id}"
}
`)

var testAccAWSRouteWithVPCEndpoint = fmt.Sprint(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-route-with-vpc-endpoint"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "terraform-testacc-route-with-vpc-endpoint"
  }
}

resource "aws_route_table" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route" "bar" {
  route_table_id         = "${aws_route_table.foo.id}"
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = "${aws_internet_gateway.foo.id}"

  # Forcing endpoint to create before route - without this the crash is a race.
  depends_on = ["aws_vpc_endpoint.baz"]
}

resource "aws_vpc_endpoint" "baz" {
  vpc_id          = "${aws_vpc.foo.id}"
  service_name    = "com.amazonaws.us-west-2.s3"
  route_table_ids = ["${aws_route_table.foo.id}"]
}
`)

var testAccAWSRouteNewRouteTable = fmt.Sprint(`
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
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-basic"
	}
}

resource "aws_internet_gateway" "bar" {
	vpc_id = "${aws_vpc.bar.id}"

	tags = {
		Name = "terraform-testacc-route-new-route-table"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-basic"
	}
}

resource "aws_route_table" "bar" {
	vpc_id = "${aws_vpc.bar.id}"

	tags = {
		Name = "terraform-testacc-route-new-route-table"
	}
}

resource "aws_route" "bar" {
	route_table_id = "${aws_route_table.bar.id}"
	destination_cidr_block = "10.4.0.0/16"
	gateway_id = "${aws_internet_gateway.bar.id}"
}
`)

func testAccAWSRouteConfigTransitGatewayIDDestinatationCidrBlock() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-route-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  route_table_id         = "${aws_vpc.test.default_route_table_id}"
  transit_gateway_id     = "${aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id}"
}
`)
}
