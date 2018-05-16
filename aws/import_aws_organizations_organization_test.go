package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func testAccAwsOrganizationsOrganization_importBasic(t *testing.T) {
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsOrganizationConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
