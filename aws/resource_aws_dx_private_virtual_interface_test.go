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
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsDxPrivateVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_basic(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_updated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "9001"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxPrivateVirtualInterface_Tags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	vpnGatewayResourceName := "aws_vpn_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_tags(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_tagsUpdated(connectionId, rName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", vpnGatewayResourceName, "id"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxPrivateVirtualInterface_DxGateway(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_private_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-private-vif-%s", acctest.RandString(9))
	amzAsn := acctest.RandIntRange(64512, 65534)
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_dxGateway(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttrSet(resourceName, "customer_address"),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", dxGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "jumbo_frame_capable", "true"),
					resource.TestCheckResourceAttr(resourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxPrivateVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckDxVirtualInterfaceDestroy(s, "aws_dx_private_virtual_interface")
}

func testAccCheckAwsDxPrivateVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckDxVirtualInterfaceExists(name, vif)
}

func testAccDxPrivateVirtualInterfaceConfig_vpnGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccDxPrivateVirtualInterfaceConfig_basic(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxPrivateVirtualInterfaceConfig_updated(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  mtu            = 9001
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxPrivateVirtualInterfaceConfig_tags(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxPrivateVirtualInterfaceConfig_tagsUpdated(cid, rName string, bgpAsn, vlan int) string {
	return testAccDxPrivateVirtualInterfaceConfig_vpnGateway(rName) + fmt.Sprintf(`
resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
  vpn_gateway_id = aws_vpn_gateway.test.id

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxPrivateVirtualInterfaceConfig_dxGateway(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  amazon_side_asn = %[3]d
  name            = %[2]q
}

resource "aws_dx_private_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[4]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[5]d
}
`, cid, rName, amzAsn, bgpAsn, vlan)
}
