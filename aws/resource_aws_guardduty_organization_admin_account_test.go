package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsGuardDutyOrganizationAdminAccount_basic(t *testing.T) {
	resourceName := "aws_guardduty_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyOrganizationAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyOrganizationAdminAccountConfigSelf(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyOrganizationAdminAccountExists(resourceName),
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

func testAccCheckAwsGuardDutyOrganizationAdminAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_organization_admin_account" {
			continue
		}

		adminAccount, err := getGuardDutyOrganizationAdminAccount(conn, rs.Primary.ID)

		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "organization is not in use") {
			continue
		}

		if err != nil {
			return err
		}

		if adminAccount == nil {
			continue
		}

		return fmt.Errorf("expected GuardDuty Organization Admin Account (%s) to be removed", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyOrganizationAdminAccountExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

		adminAccount, err := getGuardDutyOrganizationAdminAccount(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("GuardDuty Organization Admin Account (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGuardDutyOrganizationAdminAccountConfigSelf() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["guardduty.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_guardduty_detector" "test" {}

resource "aws_guardduty_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`
}
