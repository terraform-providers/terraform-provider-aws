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

func TestAccAwsDxHostedPublicVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(10))
	amazonAddress := "175.45.176.5/28"
	customerAddress := "175.45.176.6/28"
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists(resourceName, &vif),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv4"),
					resource.TestCheckResourceAttrSet(resourceName, "amazon_address"),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
			// Test import.
			{
				Config:            testAccDxHostedPublicVirtualInterfaceConfig_basic(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxHostedPublicVirtualInterface_AccepterTags(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var providers []*schema.Provider
	var vif directconnect.VirtualInterface
	resourceName := "aws_dx_hosted_public_virtual_interface.test"
	accepterResourceName := "aws_dx_hosted_public_virtual_interface_accepter.test"
	rName := fmt.Sprintf("tf-testacc-public-vif-%s", acctest.RandString(10))
	amazonAddress := "175.45.176.7/28"
	customerAddress := "175.45.176.8/28"
	bgpAsn := acctest.RandIntRange(64512, 65534)
	vlan := acctest.RandIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_accepterTags(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists(resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(connectionId, rName, amazonAddress, customerAddress, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists(resourceName, &vif),
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
					resource.TestCheckResourceAttr(resourceName, "vlan", strconv.Itoa(vlan)),
					// Accepter's attributes:
					resource.TestCheckResourceAttrSet(accepterResourceName, "arn"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(accepterResourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttrPair(accepterResourceName, "virtual_interface_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy(s *terraform.State) error {
	return testAccCheckDxVirtualInterfaceDestroy(s, "aws_dx_hosted_public_virtual_interface")
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceExists(name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return testAccCheckDxVirtualInterfaceExists(name, vif)
}

func testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Creator
resource "aws_dx_hosted_public_virtual_interface" "test" {
  address_family   = "ipv4"
  amazon_address   = %[3]q
  bgp_asn          = %[5]d
  connection_id    = %[1]q
  customer_address = %[4]q
  name             = %[2]q
  owner_account_id = data.aws_caller_identity.accepter.account_id
  vlan             = %[6]d

  route_filter_prefixes = [
    "175.45.176.0/22",
    "210.52.109.0/24",
  ]
}

# Accepter
data "aws_caller_identity" "accepter" {
  provider = "awsalternate"
}
`, cid, rName, amzAddr, custAddr, bgpAsn, vlan)
}

func testAccDxHostedPublicVirtualInterfaceConfig_basic(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + `
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id
}
`
}

func testAccDxHostedPublicVirtualInterfaceConfig_accepterTags(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccDxHostedPublicVirtualInterfaceConfig_accepterTagsUpdated(cid, rName, amzAddr, custAddr string, bgpAsn, vlan int) string {
	return testAccDxHostedPublicVirtualInterfaceConfig_base(cid, rName, amzAddr, custAddr, bgpAsn, vlan) + fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface_accepter" "test" {
  provider = "awsalternate"

  virtual_interface_id = aws_dx_hosted_public_virtual_interface.test.id

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
