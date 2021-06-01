package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterface_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_address", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_bgp_asn", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_address", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_bgp_asn", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "vlan", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterface_Id(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_address", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_bgp_asn", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_address", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_bgp_asn", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "vlan", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterface_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDataSourceName := "data.aws_ec2_local_gateway_virtual_interface.source"
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_address", sourceDataSourceName, "local_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_bgp_asn", sourceDataSourceName, "local_bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_gateway_id", sourceDataSourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_address", sourceDataSourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_bgp_asn", sourceDataSourceName, "peer_bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vlan", sourceDataSourceName, "vlan"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigFilter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  filter {
    name   = "local-gateway-virtual-interface-id"
    values = [tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]]
  }
}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigId() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  id = tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]
}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceConfigTags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "source" {
  id = tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface"
  resource_id = data.aws_ec2_local_gateway_virtual_interface.source.id
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
