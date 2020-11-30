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

func TestAccAwsDxPublicVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(10))
	// DirectConnectClientException: Amazon Address is not allowed to contain a private IP
	// DirectConnectClientException: Amazon Address and Customer Address must be in the same CIDR
	// DirectConnectClientException: Amazon Address is address 0 on its subnet.
	// DirectConnectClientException: Amazon Address is the broadcast address on its subnet.
	amazonAddress := "175.45.176.1/28"
	customerAddress := "175.45.176.2/28"
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
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

func TestAccAwsDxPublicVirtualInterface_Tags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_public_virtual_interface.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(10))
	amazonAddress := "175.45.176.3/28"
	customerAddress := "175.45.176.4/28"
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPublicVirtualInterfaceConfig_tags(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
				),
			},
			{
				Config: testAccDxPublicVirtualInterfaceConfig_tagsUpdated(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "amazon_address", amazonAddress),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_side_asn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(fmt.Sprintf("dxvif/%s", aws.StringValue(vif.VirtualInterfaceId)))),
					resource.TestCheckResourceAttrSet(resourceName, "aws_device"),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(bgpAsn)),
					resource.TestCheckResourceAttrSet(resourceName, "bgp_auth_key"),
					resource.TestCheckResourceAttr(resourceName, "connection_id", connectionId),
					resource.TestCheckResourceAttr(resourceName, "customer_address", customerAddress),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_filter_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "210.52.109.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "route_filter_prefixes.*", "175.45.176.0/22"),
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

func testAccCheckAwsDxPublicVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckDxVirtualInterfaceDestroy(s, "aws_dx_public_virtual_interface")
}

func testAccCheckAwsDxPublicVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckDxVirtualInterfaceExists(name, vif)
}

func testAccDxPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccDxPublicVirtualInterfaceConfig_tags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccDxPublicVirtualInterfaceConfig_tagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]

  tags = {
    Name = %[2]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}
