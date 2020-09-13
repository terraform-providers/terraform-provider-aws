package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_route_table", &resource.Sweeper{
		Name: "aws_route_table",
		F:    testSweepRouteTables,
	})
}

func testSweepRouteTables(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	input := &ec2.DescribeRouteTablesInput{}
	err = conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		for _, routeTable := range page.RouteTables {
			isMainRouteTableAssociation := false

			for _, routeTableAssociation := range routeTable.Associations {
				if aws.BoolValue(routeTableAssociation.Main) {
					isMainRouteTableAssociation = true
					break
				}

				input := &ec2.DisassociateRouteTableInput{
					AssociationId: routeTableAssociation.RouteTableAssociationId,
				}

				log.Printf("[DEBUG] Deleting Route Table Association: %s", input)
				_, err := conn.DisassociateRouteTable(input)
				if err != nil {
					log.Printf("[ERROR] Error deleting Route Table Association (%s): %s", aws.StringValue(routeTableAssociation.RouteTableAssociationId), err)
				}
			}

			if isMainRouteTableAssociation {
				log.Printf("[DEBUG] Skipping Main Route Table: %s", aws.StringValue(routeTable.RouteTableId))
				continue
			}

			input := &ec2.DeleteRouteTableInput{
				RouteTableId: routeTable.RouteTableId,
			}

			log.Printf("[DEBUG] Deleting Route Table: %s", input)
			_, err := conn.DeleteRouteTable(input)
			if err != nil {
				log.Printf("[ERROR] Error deleting Route Table (%s): %s", aws.StringValue(routeTable.RouteTableId), err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing Route Tables: %s", err)
	}

	return nil
}

func TestAccAWSRouteTable_basic(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_route_table.test"

	testCheck := func(*terraform.State) error {
		if len(v.Routes) != 2 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		routes := make(map[string]*ec2.Route)
		for _, r := range v.Routes {
			routes[*r.DestinationCidrBlock] = r
		}

		if _, ok := routes["10.1.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}
		if _, ok := routes["10.2.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}

	testCheckChange := func(*terraform.State) error {
		if len(v.Routes) != 3 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		routes := make(map[string]*ec2.Route)
		for _, r := range v.Routes {
			routes[*r.DestinationCidrBlock] = r
		}

		if _, ok := routes["10.1.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}
		if _, ok := routes["10.3.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}
		if _, ok := routes["10.4.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testCheck,
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteTableConfigChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testCheckChange,
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteTable_instance(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_route_table.test"

	testCheck := func(*terraform.State) error {
		if len(v.Routes) != 2 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		routes := make(map[string]*ec2.Route)
		for _, r := range v.Routes {
			routes[*r.DestinationCidrBlock] = r
		}

		if _, ok := routes["10.1.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}
		if _, ok := routes["10.2.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigInstance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testCheck,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_ipv6(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_route_table.test"

	testCheck := func(*terraform.State) error {
		// Expect 3: 2 IPv6 (local + all outbound) + 1 IPv4
		if len(v.Routes) != 3 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testCheck,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_tags(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// For GH-13545, Fixes panic on an empty route config block
func TestAccAWSRouteTable_panicEmptyRoute(t *testing.T) {
	resourceName := "aws_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRouteTableConfigPanicEmptyRoute,
				ExpectError: regexp.MustCompile("The request must contain the parameter destinationCidrBlock or destinationIpv6CidrBlock"),
			},
		},
	})
}

func TestAccAWSRouteTable_Route_ConfigMode(t *testing.T) {
	var routeTable1, routeTable2, routeTable3 ec2.RouteTable
	resourceName := "aws_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeBlocks(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeNoBlocks(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeZeroed(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable3),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_Route_TransitGatewayID(t *testing.T) {
	var routeTable1 ec2.RouteTable
	resourceName := "aws_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigRouteTransitGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_Route_LocalGatewayID(t *testing.T) {
	var routeTable1 ec2.RouteTable
	resourceName := "aws_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigRouteLocalGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRouteTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route_table" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.RouteTables) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		if !isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckRouteTableExists(n string, v *ec2.RouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.RouteTables) == 0 {
			return fmt.Errorf("RouteTable not found")
		}

		*v = *resp.RouteTables[0]

		return nil
	}
}

// VPC Peering connections are prefixed with pcx
// Right now there is no VPC Peering resource
func TestAccAWSRouteTable_vpcPeering(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_route_table.test"

	testCheck := func(*terraform.State) error {
		if len(v.Routes) != 2 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		routes := make(map[string]*ec2.Route)
		for _, r := range v.Routes {
			routes[*r.DestinationCidrBlock] = r
		}

		if _, ok := routes["10.1.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}
		if _, ok := routes["10.2.0.0/16"]; !ok {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableVpcPeeringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testCheck,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_vgwRoutePropagation(t *testing.T) {
	var v ec2.RouteTable
	var vgw ec2.VpnGateway
	resourceName := "aws_route_table.test"

	testCheck := func(*terraform.State) error {
		if len(v.PropagatingVgws) != 1 {
			return fmt.Errorf("bad propagating vgws: %#v", v.PropagatingVgws)
		}

		propagatingVGWs := make(map[string]*ec2.PropagatingVgw)
		for _, gw := range v.PropagatingVgws {
			propagatingVGWs[*gw.GatewayId] = gw
		}

		if _, ok := propagatingVGWs[*vgw.VpnGatewayId]; !ok {
			return fmt.Errorf("bad propagating vgws: %#v", v.PropagatingVgws)
		}

		return nil

	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVpnGatewayDestroy,
			testAccCheckRouteTableDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableVgwRoutePropagationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testAccCheckVpnGatewayExists("aws_vpn_gateway.test", &vgw),
					testCheck,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTable_ConditionalCidrBlock(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "0.0.0.0/0",
						"ipv6_cidr_block": "",
					}),
				),
			},
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
						"cidr_block":      "",
						"ipv6_cidr_block": "::/0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccRouteTableConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-route-table"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.2.0.0/16"
    gateway_id = aws_internet_gateway.test.id
  }
}
`

const testAccRouteTableConfigChange = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-route-table"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.3.0.0/16"
    gateway_id = aws_internet_gateway.test.id
  }

  route {
    cidr_block = "10.4.0.0/16"
    gateway_id = aws_internet_gateway.test.id
  }
}
`

const testAccRouteTableConfigIpv6 = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-route-table-ipv6"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    ipv6_cidr_block        = "::/0"
    egress_only_gateway_id = aws_egress_only_internet_gateway.test.id
  }
}
`

const testAccRouteTableConfigInstance = `
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-instance"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-route-table-instance"
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block  = "10.2.0.0/16"
    instance_id = aws_instance.test.id
  }
}
`

func testAccAWSRouteTableConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSRouteTableConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

// VPC Peering connections are prefixed with pcx
const testAccRouteTableVpcPeeringConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-vpc-peering-foo"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-route-table-vpc-peering-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-vpc-peering-bar"
  }
}

resource "aws_internet_gateway" "bar" {
  vpc_id = aws_vpc.bar.id

  tags = {
    Name = "terraform-testacc-route-table-vpc-peering-bar"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.bar.id

  tags = {
    foo = "bar"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block                = "10.2.0.0/16"
    vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  }
}
`

const testAccRouteTableVgwRoutePropagationConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-vgw-route-propagation"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id           = aws_vpc.test.id
  propagating_vgws = [aws_vpn_gateway.test.id]
}
`

// For GH-13545
const testAccRouteTableConfigPanicEmptyRoute = `
resource "aws_vpc" "test" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-route-table-panic-empty-route"
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {}
}
`

func testAccAWSRouteTableConfigRouteConfigModeBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }

  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.1.0.0/16"
    gateway_id = aws_internet_gateway.test.id
  }

  route {
    cidr_block = "10.2.0.0/16"
    gateway_id = aws_internet_gateway.test.id
  }
}
`
}

func testAccAWSRouteTableConfigRouteConfigModeNoBlocks() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }

  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}
`
}

func testAccAWSRouteTableConfigRouteConfigModeZeroed() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }
}

resource "aws_internet_gateway" "test" {
  tags = {
    Name = "tf-acc-test-ec2-route-table-config-mode"
  }

  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  route  = []
  vpc_id = aws_vpc.test.id
}
`
}

func testAccAWSRouteTableConfigRouteTransitGatewayID() string {
	return `
data "aws_availability_zones" "available" {
  # IncorrectState: Transit Gateway is not available in availability zone us-west-2d
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-table-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-route-table-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block         = "0.0.0.0/0"
    transit_gateway_id = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
  }
}
`
}

func testAccAWSRouteTableConfigRouteLocalGatewayID() string {
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

  tags = {
    Name = "tf-acc-test-ec2-route-table-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block         = "0.0.0.0/0"
    local_gateway_id = data.aws_ec2_local_gateway.first.id
  }
  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}
`
}

func testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

locals {
  ipv6             = %[2]t
  destination      = "0.0.0.0/0"
  destination_ipv6 = "::/0"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block      = local.ipv6 ? "" : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
    gateway_id      = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Route)
}
