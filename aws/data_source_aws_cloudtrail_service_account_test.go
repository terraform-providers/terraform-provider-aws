package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudTrailServiceAccount_basic(t *testing.T) {
	expectedAccountID := cloudTrailServiceAccountPerRegionMap[testAccGetRegion()]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.main", "id", expectedAccountID),
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.main", "arn", fmt.Sprintf("arn:%s:iam::%s:root", testAccGetPartition(), expectedAccountID)),
				),
			},
		},
	})
}

func TestAccAWSCloudTrailServiceAccount_Region(t *testing.T) {
	expectedAccountID := cloudTrailServiceAccountPerRegionMap[testAccGetRegion()]

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudTrailServiceAccountConfigRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.regional", "id", expectedAccountID),
					resource.TestCheckResourceAttr("data.aws_cloudtrail_service_account.regional", "arn", fmt.Sprintf("arn:%s:iam::%s:root", testAccGetPartition(), expectedAccountID)),
				),
			},
		},
	})
}

const testAccCheckAwsCloudTrailServiceAccountConfig = `
data "aws_cloudtrail_service_account" "main" { }
`

const testAccCheckAwsCloudTrailServiceAccountConfigRegion = `
data "aws_region" "current" {}

data "aws_cloudtrail_service_account" "regional" {
  region = "${data.aws_region.current.name}"
}
`
