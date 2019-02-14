package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDxPrivateVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	vifName := fmt.Sprintf("terraform-testacc-dxvif-%s", acctest.RandString(5))
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_noTags(connectionId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "tags.%", "0"),
				),
			},
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_tags(connectionId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "tags.Environment", "test"),
				),
			},
			// Test import.
			{
				ResourceName:      "aws_dx_private_virtual_interface.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxPrivateVirtualInterface_dxGateway(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	vifName := fmt.Sprintf("terraform-testacc-dxvif-%s", acctest.RandString(5))
	amzAsn := randIntRange(64512, 65534)
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_dxGateway(connectionId, vifName, amzAsn, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
				),
			},
		},
	})
}

func TestAccAwsDxPrivateVirtualInterface_mtuUpdate(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	vifName := fmt.Sprintf("terraform-testacc-dxvif-%s", acctest.RandString(5))
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxPrivateVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_noTags(connectionId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "mtu", "1500"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "jumbo_frame_capable", "true"),
				),
			},
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_jumboFrames(connectionId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "mtu", "9001"),
				),
			},
			{
				Config: testAccDxPrivateVirtualInterfaceConfig_noTags(connectionId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxPrivateVirtualInterfaceExists("aws_dx_private_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "name", vifName),
					resource.TestCheckResourceAttr("aws_dx_private_virtual_interface.foo", "mtu", "1500"),
				),
			},
		},
	})
}

func testAccCheckAwsDxPrivateVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_private_virtual_interface" {
			continue
		}

		input := &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeVirtualInterfaces(input)
		if err != nil {
			return err
		}
		for _, v := range resp.VirtualInterfaces {
			if *v.VirtualInterfaceId == rs.Primary.ID && !(*v.VirtualInterfaceState == directconnect.VirtualInterfaceStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Private VIF (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxPrivateVirtualInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxPrivateVirtualInterfaceConfig_noTags(cid, n string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "foo" {
  tags = {
    Name = "%s"
  }
}

resource "aws_dx_private_virtual_interface" "foo" {
  connection_id    = "%s"

  vpn_gateway_id = "${aws_vpn_gateway.foo.id}"
  name           = "%s"
  vlan           = %d
  address_family = "ipv4"
  bgp_asn        = %d
}
`, n, cid, n, vlan, bgpAsn)
}

func testAccDxPrivateVirtualInterfaceConfig_tags(cid, n string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "foo" {
  tags = {
    Name = "%s"
  }
}

resource "aws_dx_private_virtual_interface" "foo" {
  connection_id    = "%s"

  vpn_gateway_id = "${aws_vpn_gateway.foo.id}"
  name           = "%s"
  vlan           = %d
  address_family = "ipv4"
  bgp_asn        = %d

  tags = {
    Environment = "test"
  }
}
`, n, cid, n, vlan, bgpAsn)
}

func testAccDxPrivateVirtualInterfaceConfig_dxGateway(cid, n string, amzAsn, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "foo" {
  name            = "%s"
  amazon_side_asn = %d
}

resource "aws_dx_private_virtual_interface" "foo" {
  connection_id    = "%s"

  dx_gateway_id  = "${aws_dx_gateway.foo.id}"
  name           = "%s"
  vlan           = %d
  address_family = "ipv4"
  bgp_asn        = %d
}
`, n, amzAsn, cid, n, vlan, bgpAsn)
}

func testAccDxPrivateVirtualInterfaceConfig_jumboFrames(cid, n string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "foo" {
  tags = {
    Name = "%s"
  }
}

resource "aws_dx_private_virtual_interface" "foo" {
  connection_id    = "%s"

  vpn_gateway_id = "${aws_vpn_gateway.foo.id}"
  name           = "%s"
  vlan           = %d
  address_family = "ipv4"
  bgp_asn        = %d
  mtu            = 9001
}
`, n, cid, n, vlan, bgpAsn)
}
