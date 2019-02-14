package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	req := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
					aws.String("tf-acc-test-*"),
				},
			},
		},
	}
	resp, err := conn.DescribeRouteTables(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Route Tables: %s", err)
	}

	if len(resp.RouteTables) == 0 {
		log.Print("[DEBUG] No Route Tables to sweep")
		return nil
	}

	for _, routeTable := range resp.RouteTables {
		for _, routeTableAssociation := range routeTable.Associations {
			input := &ec2.DisassociateRouteTableInput{
				AssociationId: routeTableAssociation.RouteTableAssociationId,
			}

			log.Printf("[DEBUG] Deleting Route Table Association: %s", input)
			_, err := conn.DisassociateRouteTable(input)
			if err != nil {
				return fmt.Errorf("error deleting Route Table Association (%s): %s", aws.StringValue(routeTableAssociation.RouteTableAssociationId), err)
			}
		}

		input := &ec2.DeleteRouteTableInput{
			RouteTableId: routeTable.RouteTableId,
		}

		log.Printf("[DEBUG] Deleting Route Table: %s", input)
		_, err := conn.DeleteRouteTable(input)
		if err != nil {
			return fmt.Errorf("error deleting Route Table (%s): %s", aws.StringValue(routeTable.RouteTableId), err)
		}
	}

	return nil
}

func TestAccAWSRouteTable_basic(t *testing.T) {
	var v ec2.RouteTable

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
		IDRefreshName: "aws_route_table.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						"aws_route_table.foo", &v),
					testCheck,
					testAccCheckResourceAttrAccountID("aws_route_table.foo", "owner_id"),
				),
			},

			{
				Config: testAccRouteTableConfigChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						"aws_route_table.foo", &v),
					testCheckChange,
					testAccCheckResourceAttrAccountID("aws_route_table.foo", "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSRouteTable_instance(t *testing.T) {
	var v ec2.RouteTable

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
		IDRefreshName: "aws_route_table.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigInstance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						"aws_route_table.foo", &v),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSRouteTable_ipv6(t *testing.T) {
	var v ec2.RouteTable

	testCheck := func(*terraform.State) error {
		// Expect 3: 2 IPv6 (local + all outbound) + 1 IPv4
		if len(v.Routes) != 3 {
			return fmt.Errorf("bad routes: %#v", v.Routes)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route_table.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists("aws_route_table.foo", &v),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSRouteTable_tags(t *testing.T) {
	var route_table ec2.RouteTable

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route_table.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists("aws_route_table.foo", &route_table),
					testAccCheckTags(&route_table.Tags, "foo", "bar"),
				),
			},

			{
				Config: testAccRouteTableConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists("aws_route_table.foo", &route_table),
					testAccCheckTags(&route_table.Tags, "foo", ""),
					testAccCheckTags(&route_table.Tags, "bar", "baz"),
				),
			},
		},
	})
}

// For GH-13545, Fixes panic on an empty route config block
func TestAccAWSRouteTable_panicEmptyRoute(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_route_table.foo",
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
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidRouteTableID.NotFound" {
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
					testAccCheckRouteTableExists(
						"aws_route_table.foo", &v),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSRouteTable_vgwRoutePropagation(t *testing.T) {
	var v ec2.RouteTable
	var vgw ec2.VpnGateway

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
					testAccCheckRouteTableExists(
						"aws_route_table.foo", &v),
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &vgw),
					testCheck,
				),
			},
		},
	})
}

const testAccRouteTableConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-table"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	route {
		cidr_block = "10.2.0.0/16"
		gateway_id = "${aws_internet_gateway.foo.id}"
	}
}
`

const testAccRouteTableConfigChange = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		Name = "terraform-testacc-route-table"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	route {
		cidr_block = "10.3.0.0/16"
		gateway_id = "${aws_internet_gateway.foo.id}"
	}

	route {
		cidr_block = "10.4.0.0/16"
		gateway_id = "${aws_internet_gateway.foo.id}"
	}
}
`

const testAccRouteTableConfigIpv6 = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true
  tags = {
    Name = "terraform-testacc-route-table-ipv6"
  }
}

resource "aws_egress_only_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	route {
		ipv6_cidr_block = "::/0"
		egress_only_gateway_id = "${aws_egress_only_internet_gateway.foo.id}"
	}
}
`

const testAccRouteTableConfigInstance = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-instance"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-route-table-instance"
	}
}

resource "aws_instance" "foo" {
	# us-west-2
	ami = "ami-4fccb37f"
	instance_type = "m1.small"
	subnet_id = "${aws_subnet.foo.id}"
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	route {
		cidr_block = "10.2.0.0/16"
		instance_id = "${aws_instance.foo.id}"
	}
}
`

const testAccRouteTableConfigTags = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-tags"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		foo = "bar"
	}
}
`

const testAccRouteTableConfigTagsUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-tags"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	tags = {
		bar = "baz"
	}
}
`

// VPC Peering connections are prefixed with pcx
const testAccRouteTableVpcPeeringConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-vpc-peering-foo"
	}
}

resource "aws_internet_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

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
	vpc_id = "${aws_vpc.bar.id}"

	tags = {
		Name = "terraform-testacc-route-table-vpc-peering-bar"
	}
}

resource "aws_vpc_peering_connection" "foo" {
		vpc_id = "${aws_vpc.foo.id}"
		peer_vpc_id = "${aws_vpc.bar.id}"
	tags = {
			foo = "bar"
		}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	route {
		cidr_block = "10.2.0.0/16"
		vpc_peering_connection_id = "${aws_vpc_peering_connection.foo.id}"
	}
}
`

const testAccRouteTableVgwRoutePropagationConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-vgw-route-propagation"
	}
}

resource "aws_vpn_gateway" "foo" {
	vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

	propagating_vgws = ["${aws_vpn_gateway.foo.id}"]
}
`

// For GH-13545
const testAccRouteTableConfigPanicEmptyRoute = `
resource "aws_vpc" "foo" {
	cidr_block = "10.2.0.0/16"
	tags = {
		Name = "terraform-testacc-route-table-panic-empty-route"
	}
}

resource "aws_route_table" "foo" {
	vpc_id = "${aws_vpc.foo.id}"

  route {
  }
}
`

func testAccAWSRouteTableConfigRouteTransitGatewayID() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-route-table-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-route-table-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block         = "0.0.0.0/0"
    transit_gateway_id = "${aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id}"
  }
}
`)
}
