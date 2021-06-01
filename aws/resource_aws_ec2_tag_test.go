package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func TestAccAWSEc2Tag_basic(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2Tag_disappears(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2Tag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2Tag_Value(t *testing.T) {
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_ec2_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEc2TagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEc2TagConfig(rBgpAsn, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2TagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckEc2TagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_tag" {
			continue
		}

		resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, _, err := keyvaluetags.Ec2GetTag(conn, resourceID, key)

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Tag (%s) for resource (%s) still exists", key, resourceID)
		}
	}

	return nil
}

func testAccCheckEc2TagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		exists, _, err := keyvaluetags.Ec2GetTag(conn, resourceID, key)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Tag (%s) for resource (%s) not found", key, resourceID)
		}

		return nil
	}
}

func testAccEc2TagConfig(rBgpAsn int, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_vpn_connection.test.transit_gateway_attachment_id
  key         = %[2]q
  value       = %[3]q
}
`, rBgpAsn, key, value)
}
