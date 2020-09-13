package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGatewayRouteTable_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayRouteTableConfigLocalGatewayRouteTableId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					testAccMatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayRouteTable_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayRouteTableConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					testAccMatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayRouteTable_LocalGatewayId(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayRouteTableConfigLocalGatewayId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					testAccMatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayRouteTable_OutpostArn(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayRouteTableConfigOutpostArn(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					testAccMatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayRouteTableConfigFilter() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  filter {
    name   = "outpost-arn"
    values = [tolist(data.aws_outposts_outposts.test.arns)[0]]
  }
}
`
}

func testAccDataSourceAwsEc2LocalGatewayRouteTableConfigLocalGatewayId() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccDataSourceAwsEc2LocalGatewayRouteTableConfigLocalGatewayRouteTableId() string {
	return `
data "aws_ec2_local_gateway_route_tables" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]
}
`
}

func testAccDataSourceAwsEc2LocalGatewayRouteTableConfigOutpostArn() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}
