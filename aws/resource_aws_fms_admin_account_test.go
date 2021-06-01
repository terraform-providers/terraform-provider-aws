package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsFmsAdminAccount_basic(t *testing.T) {
	resourceName := "aws_fms_admin_account.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckFmsAdmin(t)
			testAccOrganizationsAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckFmsAdminAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFmsAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
				),
			},
		},
	})
}

func testAccCheckFmsAdminAccountDestroy(s *terraform.State) error {
	conn := testAccProviderFmsAdmin.Meta().(*AWSClient).fmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fms_admin_account" {
			continue
		}

		output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

		if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusDeleted {
			continue
		}

		return fmt.Errorf("FMS Admin Account (%s) still exists with status: %s", aws.StringValue(output.AdminAccount), aws.StringValue(output.RoleStatus))
	}

	return nil
}

func testAccFmsAdminAccountConfig_basic() string {
	return composeConfig(
		testAccFmsAdminRegionProviderConfig(),
		`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["fms.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_fms_admin_account" "test" {
  account_id = aws_organizations_organization.test.master_account_id
}
`)
}
