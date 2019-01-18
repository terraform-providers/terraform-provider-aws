package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEc2TransitGatewayRoute_basic(t *testing.T) {
	var transitGatewayRoute1 ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteConfigDestinationCidrBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayRouteExists(resourceName, &transitGatewayRoute1),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func TestAccAWSEc2TransitGatewayRoute_disappears(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayRoute1 ec2.TransitGatewayRoute
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteConfigDestinationCidrBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckAWSEc2TransitGatewayRouteExists(resourceName, &transitGatewayRoute1),
					testAccCheckAWSEc2TransitGatewayRouteDisappears(&transitGateway1, &transitGatewayRoute1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayRoute_disappears_TransitGatewayAttachment(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayRoute1 ec2.TransitGatewayRoute
	var transitGatewayVpcAttachment1 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_route.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteConfigDestinationCidrBlock(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckAWSEc2TransitGatewayRouteExists(resourceName, &transitGatewayRoute1),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(transitGatewayVpcAttachmentResourceName, &transitGatewayVpcAttachment1),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentDisappears(&transitGatewayVpcAttachment1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayRouteExists(resourceName string, transitGatewayRoute *ec2.TransitGatewayRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route ID is set")
		}

		transitGatewayRouteTableID, destination, err := decodeEc2TransitGatewayRouteID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		route, err := ec2DescribeTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if err != nil {
			return err
		}

		if route == nil {
			return fmt.Errorf("EC2 Transit Gateway Route not found")
		}

		*transitGatewayRoute = *route

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayRouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route" {
			continue
		}

		transitGatewayRouteTableID, destination, err := decodeEc2TransitGatewayRouteID(rs.Primary.ID)

		if err != nil {
			return err
		}

		route, err := ec2DescribeTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if route == nil {
			continue
		}

		return fmt.Errorf("EC2 Transit Gateway Route (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSEc2TransitGatewayRouteDisappears(transitGateway *ec2.TransitGateway, transitGatewayRoute *ec2.TransitGatewayRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteTransitGatewayRouteInput{
			DestinationCidrBlock:       transitGatewayRoute.DestinationCidrBlock,
			TransitGatewayRouteTableId: transitGateway.Options.AssociationDefaultRouteTableId,
		}

		_, err := conn.DeleteTransitGatewayRoute(input)

		return err
	}
}

func testAccAWSEc2TransitGatewayRouteConfigDestinationCidrBlock() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-route"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-route"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"
}

resource "aws_ec2_transit_gateway_route" "test" {
  destination_cidr_block         = "0.0.0.0/0"
  transit_gateway_attachment_id  = "${aws_ec2_transit_gateway_vpc_attachment.test.id}"
  transit_gateway_route_table_id = "${aws_ec2_transit_gateway.test.association_default_route_table_id}"
}
`)
}
