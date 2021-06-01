package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceAwsOrganizationsOrganizationalUnits_basic(t *testing.T) {
	resourceName := "aws_organizations_organizational_unit.test"
	dataSourceName := "data.aws_organizations_organizational_units.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsOrganizationalUnitsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "children.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "children.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "children.0.id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "children.0.arn"),
				),
			},
		},
	})
}

const testAccDataSourceAwsOrganizationsOrganizationalUnitsConfig = `
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  parent_id = aws_organizations_organization.test.roots[0].id
}

data "aws_organizations_organizational_units" "test" {
  parent_id = aws_organizations_organizational_unit.test.parent_id
}
`
