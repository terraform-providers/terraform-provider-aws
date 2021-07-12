package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceAwsOrganizationsAccounts_basic(t *testing.T) {
	resourceName := "data.aws_organizations_organization.test"
	dataSourceName := "data.aws_organizations_accounts.test1"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ErrorCheck: testAccErrorCheck(t, organizations.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsOrganizationsAccountsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "children.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.arn", dataSourceName, "children.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.email", dataSourceName, "children.0.email"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.name", dataSourceName, "children.0.name"),
					resource.TestCheckResourceAttrPair(resourceName, "non_master_accounts.0.id", dataSourceName, "children.0.id"),
				),
			},
		},
	})
}

const testAccDataSourceAwsOrganizationsAccountsConfig = `
resource "aws_organizations_organization" "test" {}

data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = "test"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test1" {
  name  = "test_account"
  email = "test1@example.com"
}

data "aws_organizations_accounts" "test1" {
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test2" {
  name      = "test_account_2"
  email     = "test2@example.com"
  parent_id = aws_organizations_organizational_unit.test.parent_id
}

data "aws_organizations_accounts" "test2" {
  parent_id = aws_organizations_organizational_unit.test.parent_id
}
`
