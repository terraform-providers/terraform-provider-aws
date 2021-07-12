package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSHostedZones_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsHostedZonesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "alb", hostedZoneIds[testAccGetRegion()]["alb"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "cloudfront", hostedZoneIds[testAccGetRegion()]["cloudfront"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "elastic_beanstalk", hostedZoneIds[testAccGetRegion()]["elastic_beanstalk"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "elb", hostedZoneIds[testAccGetRegion()]["elb"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "global_accelerator", hostedZoneIds[testAccGetRegion()]["global_accelerator"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "nlb", hostedZoneIds[testAccGetRegion()]["nlb"]),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.main", "s3", hostedZoneIds[testAccGetRegion()]["s3"]),
				),
			},
			{
				Config: testAccCheckAwsHostedZonesExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "alb", "Z32O12XQLNTSW2"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "cloudfront", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "elastic_beanstalk", "Z2NYPWQ7DFZAZH"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "elb", "Z32O12XQLNTSW2"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "global_accelerator", "Z2BJ6XQ5FK7U4H"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "nlb", "Z2IFOLAFXWLO4F"),
					resource.TestCheckResourceAttr("data.aws_hosted_zones.regional", "s3", "Z1BKCTXD74EZPE"),
				),
			},
		},
	})
}

const testAccCheckAwsHostedZonesConfig = `
data "aws_hosted_zones" "main" {}
`

//lintignore:AWSAT003
const testAccCheckAwsHostedZonesExplicitRegionConfig = `
data "aws_hosted_zones" "regional" {
  region = "eu-west-1"
}
`
