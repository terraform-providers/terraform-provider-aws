package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway_association", &resource.Sweeper{
		Name: "aws_dx_gateway_association",
		F:    testSweepDirectConnectGatewayAssociations,
	})
}

func testSweepDirectConnectGatewayAssociations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dxconn
	gatewayInput := &directconnect.DescribeDirectConnectGatewaysInput{}

	for {
		gatewayOutput, err := conn.DescribeDirectConnectGateways(gatewayInput)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Direct Connect Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Direct Connect Gateways: %s", err)
		}

		for _, gateway := range gatewayOutput.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			associationInput := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: gateway.DirectConnectGatewayId,
			}

			for {
				associationOutput, err := conn.DescribeDirectConnectGatewayAssociations(associationInput)

				if err != nil {
					return fmt.Errorf("error retrieving Direct Connect Gateway (%s) Associations: %s", directConnectGatewayID, err)
				}

				for _, association := range associationOutput.DirectConnectGatewayAssociations {
					virtualGatewayID := aws.StringValue(association.VirtualGatewayId)

					if aws.StringValue(association.AssociationState) != directconnect.GatewayAssociationStateAssociated {
						log.Printf("[INFO] Skipping Direct Connect Gateway (%s) Association in non-available (%s) state: %s", directConnectGatewayID, aws.StringValue(association.AssociationState), virtualGatewayID)
						continue
					}

					input := &directconnect.DeleteDirectConnectGatewayAssociationInput{
						DirectConnectGatewayId: gateway.DirectConnectGatewayId,
						VirtualGatewayId:       association.VirtualGatewayId,
					}

					log.Printf("[INFO] Deleting Direct Connect Gateway (%s) Association: %s", directConnectGatewayID, virtualGatewayID)
					_, err := conn.DeleteDirectConnectGatewayAssociation(input)

					if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
						continue
					}

					if err != nil {
						return fmt.Errorf("error deleting Direct Connect Gateway (%s) Association (%s): %s", directConnectGatewayID, virtualGatewayID, err)
					}

					if err := waitForDirectConnectGatewayAssociationDeletion(conn, directConnectGatewayID, virtualGatewayID, 20*time.Minute); err != nil {
						return fmt.Errorf("error waiting for Direct Connect Gateway (%s) Association (%s) to be deleted: %s", directConnectGatewayID, virtualGatewayID, err)
					}
				}

				if aws.StringValue(associationOutput.NextToken) == "" {
					break
				}

				associationInput.NextToken = associationOutput.NextToken
			}
		}

		if aws.StringValue(gatewayOutput.NextToken) == "" {
			break
		}

		gatewayInput.NextToken = gatewayOutput.NextToken
	}

	return nil
}

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
