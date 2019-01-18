package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEc2TransitGatewayDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "amazon_side_asn", dataSourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "association_default_route_table_id", dataSourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_accept_shared_attachments", dataSourceName, "auto_accept_shared_attachments"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_association", dataSourceName, "default_route_table_association"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_propagation", dataSourceName, "default_route_table_propagation"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "propagation_default_route_table_id", dataSourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway.test"
	resourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "amazon_side_asn", dataSourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "association_default_route_table_id", dataSourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_accept_shared_attachments", dataSourceName, "auto_accept_shared_attachments"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_association", dataSourceName, "default_route_table_association"),
					resource.TestCheckResourceAttrPair(resourceName, "default_route_table_propagation", dataSourceName, "default_route_table_propagation"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "propagation_default_route_table_id", dataSourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_ecmp_support", dataSourceName, "vpn_ecmp_support"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayDataSourceConfigFilter() string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

data "aws_ec2_transit_gateway" "test" {
  filter {
    name   = "transit-gateway-id"
    values = ["${aws_ec2_transit_gateway.test.id}"]
  }
}
`)
}

func testAccAWSEc2TransitGatewayDataSourceConfigID() string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

data "aws_ec2_transit_gateway" "test" {
  id = "${aws_ec2_transit_gateway.test.id}"
}
`)
}
