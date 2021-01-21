package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudTrailServiceAccount_basic(t *testing.T) {
	expectedAccountID := cloudTrailServiceAccountPerRegionMap[testAccGetRegion()]

	dataSourceName := "data.aws_cloudtrail_service_account.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					testAccCheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

func TestAccAWSCloudTrailServiceAccount_Region(t *testing.T) {
	expectedAccountID := cloudTrailServiceAccountPerRegionMap[testAccGetRegion()]

	dataSourceName := "data.aws_cloudtrail_service_account.regional"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfigRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedAccountID),
					testAccCheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", expectedAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccCheckAwsCloudTrailServiceAccountConfig = `
data "aws_cloudtrail_service_account" "main" {}
`

const testAccCheckAwsCloudTrailServiceAccountConfigRegion = `
data "aws_region" "current" {}

data "aws_cloudtrail_service_account" "regional" {
  region = data.aws_region.current.name
}
`
