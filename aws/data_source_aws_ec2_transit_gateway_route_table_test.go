package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEc2TransitGatewayRouteTableDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_table.test"
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteTableDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "default_association_route_table", dataSourceName, "default_association_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "default_propagation_route_table", dataSourceName, "default_propagation_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayRouteTableDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_table.test"
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteTableDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "default_association_route_table", dataSourceName, "default_association_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "default_propagation_route_table", dataSourceName, "default_propagation_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayRouteTableDataSourceConfigFilter() string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
}

data "aws_ec2_transit_gateway_route_table" "test" {
  filter {
    name   = "transit-gateway-route-table-id"
    values = ["${aws_ec2_transit_gateway_route_table.test.id}"]
  }
}
`)
}

func testAccAWSEc2TransitGatewayRouteTableDataSourceConfigID() string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
}

data "aws_ec2_transit_gateway_route_table" "test" {
  id = "${aws_ec2_transit_gateway_route_table.test.id}"
}
`)
}
