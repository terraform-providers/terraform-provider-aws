package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsRoute53Zones_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZonesConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRoute53ZonesDataSourceExists("data.aws_route53_zones.all"),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53ZonesDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find aws_routet53_zones data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_routet53_zones data source ID not set")
		}
		return nil
	}
}

func testAccDataSourceAwsRoute53ZonesConfig() string {
	return `
resource "aws_route53_zone" "test-zone" {
  name = "terraform-test.com."
}

data "aws_route53_zones" "all" {}
`
}
