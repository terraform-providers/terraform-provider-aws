package aws

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsDxHostedPrivateVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_private_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_private_virtual_interface_accepter.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPrivateVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			// Test import.
			{
				Config:            testAccDxHostedPrivateVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxHostedPrivateVirtualInterface_AccepterTags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_private_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_private_virtual_interface_accepter.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPrivateVirtualInterfaceConfig_accepterTags(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			{
				Config: testAccDxHostedPrivateVirtualInterfaceConfig_accepterTagsUpdated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckAwsDxHostedPrivateVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckDxVirtualInterfaceDestroy(s, "aws_dx_hosted_private_virtual_interface")
}

func testAccCheckAwsDxHostedPrivateVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckDxVirtualInterfaceExists(name, vif)
}

func testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName string, bgpAsn, vlan int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_private_virtual_interface" "test" {
  address_family   = "ipv4"
  bgp_asn          = %[3]d
  connection_id    = %[1]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[4]d

  # The aws_dx_hosted_private_virtual_interface
  # must be destroyed before the aws_vpn_gateway.
  depends_on = [aws_vpn_gateway.test]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}

resource "aws_vpn_gateway" "test" {
  provider = "awsalternate"

  tags = {
    Name = %[2]q
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxHostedPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + `
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id
}
`
}

func testAccDxHostedPrivateVirtualInterfaceConfig_accepterTags(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccDxHostedPrivateVirtualInterfaceConfig_accepterTagsUpdated(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxHostedPrivateVirtualInterfaceConfig_base(cid, rName, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_private_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_private_virtual_interface.test.id
  vpn_gateway_id       = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
