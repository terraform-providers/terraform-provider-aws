package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsVpnGateway_unattached(t *testing.T) {
	rInt := acctest.RandInt()
	dataSourceNameById := "data.aws_vpn_gateway.test_by_id"
	dataSourceNameByTags := "data.aws_vpn_gateway.test_by_tags"
	dataSourceNameByAsn := "data.aws_vpn_gateway.test_by_amazon_side_asn"
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpnGatewayUnattachedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameById, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameById, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByAsn, "id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "state"),
					resource.TestCheckResourceAttr(dataSourceNameByTags, "tags.%", "3"),
					resource.TestCheckNoResourceAttr(dataSourceNameById, "attached_vpc_id"),
					resource.TestCheckResourceAttr(dataSourceNameByAsn, "amazon_side_asn", "4294967293"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpnGateway_attached(t *testing.T) {
	rInt := acctest.RandInt()
	dataSourceName := "data.aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpnGatewayAttachedConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "aws_vpn_gateway.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attached_vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(dataSourceName, "state", regexp.MustCompile("(?i)available")),
				),
			},
		},
	})
}

func testAccDataSourceAwsVpnGatewayUnattachedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-unattached-%d"
    ABC  = "testacc-%d"
    XYZ  = "testacc-%d"
  }

  amazon_side_asn = 4294967293
}

data "aws_vpn_gateway" "test_by_id" {
  id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test_by_tags" {
  tags = aws_vpn_gateway.test.tags
}

data "aws_vpn_gateway" "test_by_amazon_side_asn" {
  amazon_side_asn = aws_vpn_gateway.test.amazon_side_asn
  state           = "available"
}
`, rInt, rInt+1, rInt-1)
}

func testAccDataSourceAwsVpnGatewayAttachedConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-attached-%d"
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-attached-%d"
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test" {
  attached_vpc_id = aws_vpn_gateway_attachment.test.vpc_id
}
`, rInt, rInt)
}
