package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCustomerGatewayDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"

	asn := acctest.RandIntRange(64512, 65534)
	hostOctet := acctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCustomerGatewayDataSourceConfigFilter(asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_address", dataSourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSCustomerGatewayDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"

	asn := acctest.RandIntRange(64512, 65534)
	hostOctet := acctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCustomerGatewayDataSourceConfigID(asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_address", dataSourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
				),
			},
		},
	})
}

func testAccAWSCustomerGatewayDataSourceConfigFilter(asn, hostOctet int) string {
	name := acctest.RandomWithPrefix("test-filter")
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "50.0.0.%d"
  type       = "ipsec.1"

  tags = {
    Name = "%s"
  }
}

data "aws_customer_gateway" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_customer_gateway.test.tags.Name]
  }
}
`, asn, hostOctet, name)
}

func testAccAWSCustomerGatewayDataSourceConfigID(asn, hostOctet int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "50.0.0.%d"
  type       = "ipsec.1"
}

data "aws_customer_gateway" "test" {
  id = aws_customer_gateway.test.id
}
`, asn, hostOctet)
}
