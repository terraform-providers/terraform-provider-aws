package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSElbHostedZoneId_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElbHostedZoneIdConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.main", "id", elbHostedZoneIdPerRegionMap[testAccGetRegion()]),
				),
			},
			{
				Config: testAccCheckAwsElbHostedZoneIdExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "Z32O12XQLNTSW2"),
				),
			},
		},
	})
}

const testAccCheckAwsElbHostedZoneIdConfig = `
data "aws_elb_hosted_zone_id" "main" {}
`

//lintignore:AWSAT003
const testAccCheckAwsElbHostedZoneIdExplicitRegionConfig = `
data "aws_elb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`
