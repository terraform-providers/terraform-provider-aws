package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

func TestAccAWSServiceCatalogOrganizationsAccess_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_organizations_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsEnabledPreCheck(t)
			testAccOrganizationManagementAccountPreCheck(t)
		},
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogOrganizationsAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogOrganizationsAccessConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogOrganizationsAccessExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceCatalogOrganizationsAccessDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_organizations_access" {
			continue
		}

		output, err := waiter.OrganizationsAccessStable(conn)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog AWS Organizations Access (%s): %w", rs.Primary.ID, err)
		}

		if output == "" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): empty response", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckAwsServiceCatalogOrganizationsAccessExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		output, err := waiter.OrganizationsAccessStable(conn)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog AWS Organizations Access (%s): %w", rs.Primary.ID, err)
		}

		if output == "" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): empty response", rs.Primary.ID)
		}

		if output != servicecatalog.AccessStatusEnabled && rs.Primary.Attributes["enabled"] == "true" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		if output == servicecatalog.AccessStatusEnabled && rs.Primary.Attributes["enabled"] == "false" {
			return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSServiceCatalogOrganizationsAccessConfig_basic() string {
	return `
resource "aws_servicecatalog_organizations_access" "test" {
  enabled = "true"
}
`
}
