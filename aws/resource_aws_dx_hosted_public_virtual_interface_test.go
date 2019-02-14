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

func TestAccAwsDxHostedPublicVirtualInterface_basic(t *testing.T) {
	key := "DX_CONNECTION_ID"
	connectionId := os.Getenv(key)
	if connectionId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	key = "DX_HOSTED_VIF_OWNER_ACCOUNT"
	ownerAccountId := os.Getenv(key)
	if ownerAccountId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	vifName := fmt.Sprintf("terraform-testacc-dxvif-%s", acctest.RandString(5))
	bgpAsn := randIntRange(64512, 65534)
	vlan := randIntRange(2049, 4094)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedPublicVirtualInterfaceConfig_basic(connectionId, ownerAccountId, vifName, bgpAsn, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedPublicVirtualInterfaceExists("aws_dx_hosted_public_virtual_interface.foo"),
					resource.TestCheckResourceAttr("aws_dx_hosted_public_virtual_interface.foo", "name", vifName),
				),
			},
			// Test import.
			{
				ResourceName:      "aws_dx_hosted_public_virtual_interface.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_hosted_public_virtual_interface" {
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
				return fmt.Errorf("[DESTROY ERROR] Dx Public VIF (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxHostedPublicVirtualInterfaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxHostedPublicVirtualInterfaceConfig_basic(cid, ownerAcctId, n string, bgpAsn, vlan int) string {
	return fmt.Sprintf(`
resource "aws_dx_hosted_public_virtual_interface" "foo" {
  connection_id    = "%s"
  owner_account_id = "%s"

  name           = "%s"
  vlan           = %d
  address_family = "ipv4"
  bgp_asn        = %d

  customer_address = "175.45.176.1/30"
  amazon_address   = "175.45.176.2/30"
  route_filter_prefixes = [
    "210.52.109.0/24",
	"175.45.176.0/22"
  ]
}
`, cid, ownerAcctId, n, vlan, bgpAsn)
}
