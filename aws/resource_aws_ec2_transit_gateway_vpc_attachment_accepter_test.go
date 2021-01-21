package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSEc2TransitGatewayVpcAttachmentAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	vpcAttachmentName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := fmt.Sprintf("tf-testacc-tgwvpcattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayVpcAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", vpcAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", callerIdentityDatasourceName, "account_id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayVpcAttachmentAccepter_Tags(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	vpcAttachmentName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := fmt.Sprintf("tf-testacc-tgwvpcattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayVpcAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", vpcAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", callerIdentityDatasourceName, "account_id"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", vpcAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", callerIdentityDatasourceName, "account_id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_tagsUpdated(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayVpcAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation(t *testing.T) {
	var providers []*schema.Provider
	var transitGateway ec2.TransitGateway
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := fmt.Sprintf("tf-testacc-tgwvpcattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayVpcAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentNotAssociated(&transitGateway, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentNotPropagated(&transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentAssociated(&transitGateway, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentNotPropagated(&transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentNotAssociated(&transitGateway, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentPropagated(&transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(transitGatewayResourceName, &transitGateway),
					testAccCheckAWSEc2TransitGatewayVpcAttachmentExists(resourceName, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentAssociated(&transitGateway, &transitGatewayVpcAttachment),
					testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentPropagated(&transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_base(rName string) string {
	return composeConfig(testAccAlternateAccountProviderConfig(), testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.creator.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

# VPC attachment creator.
data "aws_caller_identity" "creator" {
  provider = "awsalternate"
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  provider = "awsalternate"

  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
    Side = "Creator"
  }
}
`, rName))
}

func testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_basic(rName string) string {
	return testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_base(rName) + `
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`
}

func testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_tags(rName string) string {
	return testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_tagsUpdated(rName string) string {
	return testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key3 = "Value3"
    Key2 = "Value2b"
  }
}
`, rName)
}

func testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName string, association, propagation bool) string {
	return testAccAWSEc2TransitGatewayVpcAttachmentAccepterConfig_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
    Side = "Accepter"
  }

  transit_gateway_default_route_table_association = %[2]t
  transit_gateway_default_route_table_propagation = %[3]t
}
`, rName, association, propagation)
}
