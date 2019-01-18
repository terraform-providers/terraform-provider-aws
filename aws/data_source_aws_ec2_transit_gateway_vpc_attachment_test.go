package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEc2TransitGatewayVpcAttachmentDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_vpc_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_support", dataSourceName, "ipv6_support"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", dataSourceName, "vpc_owner_id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayVpcAttachmentDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_vpc_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_support", dataSourceName, "ipv6_support"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", dataSourceName, "vpc_owner_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayVpcAttachmentDataSourceConfigFilter() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"
}

data "aws_ec2_transit_gateway_vpc_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["${aws_ec2_transit_gateway_vpc_attachment.test.id}"]
  }
}
`)
}

func testAccAWSEc2TransitGatewayVpcAttachmentDataSourceConfigID() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"
}

data "aws_ec2_transit_gateway_vpc_attachment" "test" {
  id = "${aws_ec2_transit_gateway_vpc_attachment.test.id}"
}
`)
}
