package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxGatewayAssociation_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig(acctest.RandString(5), randIntRange(64512, 65534)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists("aws_dx_gateway_association.test"),
				),
			},
		},
	})
}

func TestAccAwsDxGatewayAssociation_multiVgws(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVgws(acctest.RandString(5), randIntRange(64512, 65534)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationExists("aws_dx_gateway_association.test1"),
					testAccCheckAwsDxGatewayAssociationExists("aws_dx_gateway_association.test2"),
				),
			},
		},
	})
}

func testAccCheckAwsDxGatewayAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association" {
			continue
		}

		resp, _ := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			DirectConnectGatewayId: aws.String(rs.Primary.Attributes["dx_gateway_id"]),
			VirtualGatewayId:       aws.String(rs.Primary.Attributes["vpn_gateway_id"]),
		})

		if len(resp.DirectConnectGatewayAssociations) > 0 {
			return fmt.Errorf("Direct Connect Gateway (%s) is not dissociated from VGW %s", rs.Primary.Attributes["dx_gateway_id"], rs.Primary.Attributes["vpn_gateway_id"])
		}
	}
	return nil
}

func testAccCheckAwsDxGatewayAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxGatewayAssociationConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name = "terraform-testacc-dxgwassoc-%s"
  amazon_side_asn = "%d"
}

resource "aws_vpc" "test" {
  cidr_block = "10.255.255.0/28"
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s"
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s"
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id = "${aws_vpc.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test.id}"
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id = "${aws_dx_gateway.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway_attachment.test.vpn_gateway_id}"
}
`, rName, rBgpAsn, rName, rName)
}

func testAccDxGatewayAssociationConfig_multiVgws(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name = "terraform-testacc-dxgwassoc-%s"
  amazon_side_asn = "%d"
}

resource "aws_vpc" "test1" {
  cidr_block = "10.255.255.16/28"
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s-1"
  }
}

resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s-1"
  }
}

resource "aws_vpn_gateway_attachment" "test1" {
  vpc_id = "${aws_vpc.test1.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test1.id}"
}

resource "aws_dx_gateway_association" "test1" {
  dx_gateway_id = "${aws_dx_gateway.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway_attachment.test1.vpn_gateway_id}"
}

resource "aws_vpc" "test2" {
  cidr_block = "10.255.255.32/28"
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s-2"
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = "terraform-testacc-dxgwassoc-%s-2"
  }
}

resource "aws_vpn_gateway_attachment" "test2" {
  vpc_id = "${aws_vpc.test2.id}"
  vpn_gateway_id = "${aws_vpn_gateway.test2.id}"
}

resource "aws_dx_gateway_association" "test2" {
  dx_gateway_id = "${aws_dx_gateway.test.id}"
  vpn_gateway_id = "${aws_vpn_gateway_attachment.test2.vpn_gateway_id}"
}
`, rName, rBgpAsn, rName, rName, rName, rName)
}
