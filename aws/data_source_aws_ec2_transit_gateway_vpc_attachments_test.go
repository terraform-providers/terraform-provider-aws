package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEc2TransitGatewayVpcAttachmentsDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_vpc_attachments.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentsDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayVpcAttachmentsDataSourceConfigFilter() string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # IncorrectState: Transit Gateway is not available in availability zone us-west-2d
  blacklisted_zone_ids = ["usw2-az4"]
  state                = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test2" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.1.0/24"
  vpc_id            = aws_vpc.test2.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test2" {
  subnet_ids         = [aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test2.id
}

data "aws_ec2_transit_gateway_vpc_attachments" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_vpc_attachments.test.id]
  }
}
`)
}
