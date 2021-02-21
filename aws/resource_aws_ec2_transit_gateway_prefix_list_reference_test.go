package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func TestAccAwsEc2TransitGatewayPrefixListReference_basic(t *testing.T) {
	managedPrefixListResourceName := "aws_ec2_managed_prefix_list.test"
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2TransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2TransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", managedPrefixListResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "prefix_list_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayResourceName, "association_default_route_table_id"),
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

func TestAccAwsEc2TransitGatewayPrefixListReference_disappears(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2TransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2TransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2TransitGatewayPrefixListReference(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2TransitGatewayPrefixListReference_disappears_TransitGateway(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2TransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2TransitGatewayPrefixListReferenceConfig_Blackhole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2TransitGateway(), transitGatewayResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2TransitGatewayPrefixListReference_TransitGatewayAttachmentId(t *testing.T) {
	resourceName := "aws_ec2_transit_gateway_prefix_list_reference.test"
	transitGatewayVpcAttachmentResourceName1 := "aws_ec2_transit_gateway_vpc_attachment.test.0"
	transitGatewayVpcAttachmentResourceName2 := "aws_ec2_transit_gateway_vpc_attachment.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccPreCheckEc2ManagedPrefixList(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2TransitGatewayPrefixListReferenceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2TransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentId(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEc2TransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentId(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "blackhole", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName2, "id"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2TransitGatewayPrefixListReferenceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_prefix_list_reference" {
			continue
		}

		transitGatewayPrefixListReference, err := finder.TransitGatewayPrefixListReferenceByID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading EC2 Transit Gateway Prefix List Reference (%s): %w", rs.Primary.ID, err)
		}

		if transitGatewayPrefixListReference != nil {
			return fmt.Errorf("EC2 Transit Gateway Prefix List Reference (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsEc2TransitGatewayPrefixListReferenceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		transitGatewayPrefixListReference, err := finder.TransitGatewayPrefixListReferenceByID(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading EC2 Transit Gateway Prefix List Reference (%s): %w", rs.Primary.ID, err)
		}

		if transitGatewayPrefixListReference == nil {
			return fmt.Errorf("EC2 Transit Gateway Prefix List Reference (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsEc2TransitGatewayPrefixListReferenceConfig_Blackhole(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  blackhole                      = true
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName)
}

func testAccAwsEc2TransitGatewayPrefixListReferenceConfig_TransitGatewayAttachmentId(rName string, index int) string {
	return fmt.Sprintf(`
variable "index" {
  default = %[2]d
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_vpc" "test" {
  count = 2

  cidr_block = "10.${count.index}.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  cidr_block = cidrsubnet(aws_vpc.test[count.index].cidr_block, 8, 0)
  vpc_id     = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  count = 2

  subnet_ids         = [aws_subnet.test[count.index].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test[count.index].id
}

resource "aws_ec2_transit_gateway_prefix_list_reference" "test" {
  prefix_list_id                 = aws_ec2_managed_prefix_list.test.id
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test[var.index].id
  transit_gateway_route_table_id = aws_ec2_transit_gateway.test.association_default_route_table_id
}
`, rName, index)
}
