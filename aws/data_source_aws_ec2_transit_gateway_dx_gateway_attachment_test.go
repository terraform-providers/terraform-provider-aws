package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2TransitGatewayDxGatewayAttachmentDataSource_TransitGatewayIdAndDxGatewayId(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	dataSourceName := "data.aws_ec2_transit_gateway_dx_gateway_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	dxGatewayResourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDxAttachmentDataSourceConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dx_gateway_id", dxGatewayResourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayDxGatewayAttachmentDataSource_filter(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	dataSourceName := "data.aws_ec2_transit_gateway_dx_gateway_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	dxGatewayResourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayDxAttachmentDataSourceConfigFilter(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dx_gateway_id", dxGatewayResourceName, "id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayDxAttachmentDataSourceConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

data "aws_ec2_transit_gateway_dx_gateway_attachment" "test" {
  transit_gateway_id = aws_dx_gateway_association.test.associated_gateway_id
  dx_gateway_id      = aws_dx_gateway_association.test.dx_gateway_id
}
`, rName, rBgpAsn)
}

func testAccAWSEc2TransitGatewayDxAttachmentDataSourceConfigFilter(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}

data "aws_ec2_transit_gateway_dx_gateway_attachment" "test" {
  filter {
    name   = "resource-id"
    values = [aws_dx_gateway_association.test.dx_gateway_id]
  }
}
`, rName, rBgpAsn)
}
