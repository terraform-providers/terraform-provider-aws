package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccAWSGuarddutyDetectorDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuarddutyDetectorBasicResourceConfig(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccAwsGuarddutyDetectorBasicResourceDataConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					testAccCheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
				),
			},
		},
	})
}

func testAccAWSGuarddutyDetectorDataSource_Id(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuarddutyDetectorExplicitConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					testAccCheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
				),
			},
		},
	})
}

func testAccAwsGuarddutyDetectorBasicResourceConfig() string {
	return `
resource "aws_guardduty_detector" "test" {}
`
}

func testAccAwsGuarddutyDetectorBasicResourceDataConfig() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {}
`
}

func testAccAwsGuarddutyDetectorExplicitConfig() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {
  id = aws_guardduty_detector.test.id
}
`
}
