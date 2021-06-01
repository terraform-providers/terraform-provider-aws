package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSLakeFormationDataLakeSettings_basic(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationDataLakeSettingsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "catalog_id", callerIdentityName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", callerIdentityName, "arn"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationDataLakeSettings_disappears(t *testing.T) {
	resourceName := "aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationDataLakeSettingsExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLakeFormationDataLakeSettings(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSLakeFormationDataLakeSettings_withoutCatalogId(t *testing.T) {
	callerIdentityName := "data.aws_caller_identity.current"
	resourceName := "aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationDataLakeSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationDataLakeSettingsConfig_withoutCatalogId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationDataLakeSettingsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", callerIdentityName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAWSLakeFormationDataLakeSettingsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_data_lake_settings" {
			continue
		}

		input := &lakeformation.GetDataLakeSettingsInput{}

		if rs.Primary.Attributes["catalog_id"] != "" {
			input.CatalogId = aws.String(rs.Primary.Attributes["catalog_id"])
		}

		output, err := conn.GetDataLakeSettings(input)

		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Lake Formation data lake settings (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && output.DataLakeSettings != nil && len(output.DataLakeSettings.DataLakeAdmins) > 0 {
			return fmt.Errorf("Lake Formation data lake admin(s) (%s) still exist", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSLakeFormationDataLakeSettingsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

		input := &lakeformation.GetDataLakeSettingsInput{}

		if rs.Primary.Attributes["catalog_id"] != "" {
			input.CatalogId = aws.String(rs.Primary.Attributes["catalog_id"])
		}

		_, err := conn.GetDataLakeSettings(input)

		if err != nil {
			return fmt.Errorf("error getting Lake Formation data lake settings (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

const testAccAWSLakeFormationDataLakeSettingsConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = data.aws_caller_identity.current.account_id

  create_database_default_permissions {
    principal   = "IAM_ALLOWED_PRINCIPALS"
    permissions = ["ALL"]
  }

  create_table_default_permissions {
    principal   = "IAM_ALLOWED_PRINCIPALS"
    permissions = ["ALL"]
  }

  admins                  = [data.aws_caller_identity.current.arn]
  trusted_resource_owners = [data.aws_caller_identity.current.account_id]
}
`

const testAccAWSLakeFormationDataLakeSettingsConfig_withoutCatalogId = `
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}
`
