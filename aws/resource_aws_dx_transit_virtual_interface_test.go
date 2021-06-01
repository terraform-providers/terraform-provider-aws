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

func TestAccAwsDxTransitVirtualInterface_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic": testAccAwsDxTransitVirtualInterface_basic,
		"tags":  testAccAwsDxTransitVirtualInterface_Tags,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsDxTransitVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_transit_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", acctest.RandString(9))
	amzAsn := acctest.RandIntRange(64512, 65534)
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxTransitVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxTransitVirtualInterfaceConfig_basic(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName, &vif),
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
			{
				Config: testAccDxTransitVirtualInterfaceConfig_updated(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "mtu", "8500"),
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

func testAccAwsDxTransitVirtualInterface_Tags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_transit_virtual_interface.test"
	dxGatewayResourceName := "aws_dx_gateway.test"
	rName := fmt.Sprintf("tf-testacc-transit-vif-%s", acctest.RandString(9))
	amzAsn := acctest.RandIntRange(64512, 65534)
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxTransitVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxTransitVirtualInterfaceConfig_tags(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			{
				Config: testAccDxTransitVirtualInterfaceConfig_tagsUpdated(connectionId, rName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxTransitVirtualInterfaceExists(resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
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

func testAccCheckAwsDxTransitVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckDxVirtualInterfaceExists(name, vif)
}

func testAccCheckAwsDxTransitVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckDxVirtualInterfaceDestroy(s, "aws_dx_transit_virtual_interface")
}

func testAccDxTransitVirtualInterfaceConfig_base(rName string, amzAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = %[2]d
}
`, rName, amzAsn)
}

func testAccDxTransitVirtualInterfaceConfig_basic(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxTransitVirtualInterfaceConfig_base(rName, amzAsn) + fmt.Sprintf(`
resource "aws_dx_transit_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxTransitVirtualInterfaceConfig_updated(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxTransitVirtualInterfaceConfig_base(rName, amzAsn) + fmt.Sprintf(`
resource "aws_dx_transit_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  mtu            = 8500
  name           = %[2]q
  vlan           = %[4]d
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxTransitVirtualInterfaceConfig_tags(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxTransitVirtualInterfaceConfig_base(rName, amzAsn) + fmt.Sprintf(`
resource "aws_dx_transit_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, bgpAsn, vlan)
}

func testAccDxTransitVirtualInterfaceConfig_tagsUpdated(cid, rName string, amzAsn, bgpAsn, vlan int) string {
	return testAccDxTransitVirtualInterfaceConfig_base(rName, amzAsn) + fmt.Sprintf(`
resource "aws_dx_transit_virtual_interface" "test" {
  address_family = "ipv4"
  bgp_asn        = %[3]d
  dx_gateway_id  = aws_dx_gateway.test.id
  connection_id  = %[1]q
  name           = %[2]q
  vlan           = %[4]d

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, bgpAsn, vlan)
}
