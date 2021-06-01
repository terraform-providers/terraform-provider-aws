package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2TransitGatewayVpnAttachmentDataSource_TransitGatewayIdAndVpnConnectionId(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	dataSourceName := "data.aws_ec2_transit_gateway_vpn_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpnConnectionResourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigTransitGatewayIdAndVpnConnectionId(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_connection_id", vpnConnectionResourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayVpnAttachmentDataSource_filter(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	dataSourceName := "data.aws_ec2_transit_gateway_vpn_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpnConnectionResourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigFilter(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_connection_id", vpnConnectionResourceName, "id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigBase(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-transit-gateway-id"
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-transit-gateway-id"
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type

  tags = {
    Name = "tf-acc-test-ec2-vpn-connection-transit-gateway-id"
  }
}
`, rBgpAsn)
}

func testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigTransitGatewayIdAndVpnConnectionId(rBgpAsn int) string {
	return testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigBase(rBgpAsn) + `
data "aws_ec2_transit_gateway_vpn_attachment" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpn_connection_id  = aws_vpn_connection.test.id
}
`
}

func testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigFilter(rBgpAsn int) string {
	return testAccAWSEc2TransitGatewayVpnAttachmentDataSourceConfigBase(rBgpAsn) + `
data "aws_ec2_transit_gateway_vpn_attachment" "test" {
  filter {
    name   = "resource-id"
    values = [aws_vpn_connection.test.id]
  }
}
`
}
