package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEc2TransitGatewayRouteTableAssociation_basic(t *testing.T) {
	var transitGatewayRouteTablePropagtion1 ec2.TransitGatewayRouteTableAssociation
	resourceName := "aws_ec2_transit_gateway_route_table_association.test"
	transitGatewayRouteTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteTableAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayRouteTableAssociationExists(resourceName, &transitGatewayRouteTablePropagtion1),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayRouteTableResourceName, "id"),
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

func testAccCheckAWSEc2TransitGatewayRouteTableAssociationExists(resourceName string, transitGatewayRouteTableAssociation *ec2.TransitGatewayRouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route ID is set")
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := decodeEc2TransitGatewayRouteTableAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		association, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if err != nil {
			return err
		}

		if association == nil {
			return fmt.Errorf("EC2 Transit Gateway Route Table Association not found")
		}

		*transitGatewayRouteTableAssociation = *association

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayRouteTableAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table_association" {
			continue
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := decodeEc2TransitGatewayRouteTableAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		association, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if association == nil {
			continue
		}

		return fmt.Errorf("EC2 Transit Gateway Route Table (%s) Association (%s) still exists", transitGatewayRouteTableID, transitGatewayAttachmentID)
	}

	return nil
}

func testAccAWSEc2TransitGatewayRouteTableAssociationConfig() string {
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
  subnet_ids                                      = ["${aws_subnet.test.id}"]
  transit_gateway_default_route_table_association = false
  transit_gateway_id                              = "${aws_ec2_transit_gateway.test.id}"
  vpc_id                                          = "${aws_vpc.test.id}"
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = "${aws_ec2_transit_gateway_vpc_attachment.test.id}"
  transit_gateway_route_table_id = "${aws_ec2_transit_gateway_route_table.test.id}"
}
`)
}
