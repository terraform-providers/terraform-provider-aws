package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEc2CoipPool_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2CoipPoolDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					resource.TestMatchResourceAttr(dataSourceName, "pool_id", regexp.MustCompile(`^ipv4pool-coip-`)),
					testCheckResourceAttrGreaterThanValue(dataSourceName, "pool_cidrs.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2CoipPool_Id(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2CoipPoolDataSourceConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					resource.TestMatchResourceAttr(dataSourceName, "pool_id", regexp.MustCompile(`^ipv4pool-coip-`)),
					testCheckResourceAttrGreaterThanValue(dataSourceName, "pool_cidrs.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2CoipPoolDataSourceConfigFilter() string {
	return `
data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_coip_pool" "test" {
  filter {
    name   = "coip-pool.pool-id"
    values = [tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]]
  }
}
`
}

func testAccDataSourceAwsEc2CoipPoolDataSourceConfigId() string {
	return `
data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}
`
}
