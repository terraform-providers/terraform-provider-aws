package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, "owner_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_ID_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_ID_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, "owner_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Tags(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigTags_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigTags_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1_sameAccount(rName, "Name", rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  tags = {
    Name = aws_ec2_transit_gateway_peering_attachment.test.tags["Name"]
  }
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter_differentAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"

  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID_differentAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"
  id       = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}
