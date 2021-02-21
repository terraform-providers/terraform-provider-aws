package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
)

func testAccAwsSecurityHubOrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityHubOrganizationAdminAccountConfigSelf(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubOrganizationAdminAccountExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "admin_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsSecurityHubOrganizationAdminAccount_disappears(t *testing.T) {
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityHubOrganizationAdminAccountConfigSelf(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubOrganizationAdminAccountExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSecurityHubOrganizationAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsSecurityHubOrganizationAdminAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_organization_admin_account" {
			continue
		}

		adminAccount, err := finder.AdminAccount(conn, rs.Primary.ID)

		// Because of this resource's dependency, the Organizations organization
		// will be deleted first, resulting in the following valid error
		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeAccessDeniedException, "account is not a member of an organization") {
			continue
		}

		if err != nil {
			return err
		}

		if adminAccount == nil {
			continue
		}

		return fmt.Errorf("expected Security Hub Organization Admin Account (%s) to be removed", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsSecurityHubOrganizationAdminAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		adminAccount, err := finder.AdminAccount(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("Security Hub Organization Admin Account (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSecurityHubOrganizationAdminAccountConfigSelf() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`
}
