package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccAwsGuardDutyOrganizationConfiguration_basic(t *testing.T) {
	detectorResourceName := "aws_guardduty_detector.test"
	resourceName := "aws_guardduty_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		Providers: testAccProviders,
		// GuardDuty Organization Configuration cannot be deleted separately.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckAwsGuardDutyDetectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigAutoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyOrganizationConfigurationConfigAutoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
				),
			},
		},
	})
}

func testAccGuardDutyOrganizationConfigurationConfigAutoEnable(autoEnable bool) string {
	return fmt.Sprintf(`
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

resource "aws_guardduty_organization_configuration" "test" {
  depends_on = [aws_guardduty_organization_admin_account.test]

  auto_enable = %[1]t
  detector_id = aws_guardduty_detector.test.id
}
`, autoEnable)
}
