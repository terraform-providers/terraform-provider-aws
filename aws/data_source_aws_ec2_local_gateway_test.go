package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGateway_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-`)),
					testAccMatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					testAccCheckResourceAttrAccountID(dataSourceName, "owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayConfigId() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway" "test" {
  id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}
