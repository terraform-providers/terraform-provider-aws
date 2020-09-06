package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroups_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroups_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroups_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfig() string {
	return `
data "aws_ec2_local_gateway_virtual_interface_groups" "test" {}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigFilter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigTags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_groups"
  resource_id = tolist(data.aws_ec2_local_gateway_virtual_interface_groups.source.ids)[0]
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
