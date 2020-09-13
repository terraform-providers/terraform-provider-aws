package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsRegions_basic(t *testing.T) {
	resourceName := "data.aws_regions.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegions_Filter(t *testing.T) {
	resourceName := "data.aws_regions.opt_in_status"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegionsFiltered("opt-in-not-required"),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegions_AllRegions(t *testing.T) {
	resourceAllRegions := "data.aws_regions.all_regions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceAllRegions),
					resource.TestCheckResourceAttr(resourceAllRegions, "all_regions", "true"),
					resource.TestCheckNoResourceAttr(resourceAllRegions, "opt_in_status"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRegionsCheck(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		names, namesOk := rs.Primary.Attributes["names.#"]

		if !namesOk {
			return fmt.Errorf("names attribute is missing.")
		}

		namesQuantity, err := strconv.Atoi(names)

		if err != nil {
			return fmt.Errorf("error parsing names (%s) into integer: %s", names, err)
		}

		if namesQuantity == 0 {
			return fmt.Errorf("No names found, this is probably a bug.")
		}

		return nil
	}
}

func testAccDataSourceAwsRegionsConfig_empty() string {
	return `
data "aws_regions" "empty" {}
`
}

func testAccDataSourceAwsRegionsConfig_allRegions() string {
	return `
data "aws_regions" "all_regions" {
  all_regions = "true"
}
`
}

func testAccDataSourceAwsRegionsConfig_allRegionsFiltered(filter string) string {
	return fmt.Sprintf(`
data "aws_regions" "opt_in_status" {
  filter {
    name   = "opt-in-status"
    values = ["%s"]
  }
}
`, filter)
}
