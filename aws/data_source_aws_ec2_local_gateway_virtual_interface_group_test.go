package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroup_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroup_LocalGatewayId(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigLocalGatewayId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroup_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	sourceDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.source"
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_gateway_id", sourceDataSourceName, "local_gateway_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigFilter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigLocalGatewayId() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupConfigTags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_group"
  resource_id = data.aws_ec2_local_gateway_virtual_interface_group.source.id
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
