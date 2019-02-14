package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceElasticBeanstalkHostedZone(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elastic_beanstalk_hosted_zone.current", "id", "Z2PCDNR3VC2G1N"),
				),
			},
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("ap-southeast-2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elastic_beanstalk_hosted_zone.test", "id", "Z2PCDNR3VC2G1N"),
				),
			},
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("eu-west-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elastic_beanstalk_hosted_zone.test", "id", "Z2NYPWQ7DFZAZH"),
				),
			},
			{
				Config:      testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("ss-pluto-1"),
				ExpectError: regexp.MustCompile("Unsupported region"),
			},
		},
	})
}

const testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion = `
provider "aws" {
	region = "ap-southeast-2"
}
data "aws_elastic_beanstalk_hosted_zone" "current" {}
`

func testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion(r string) string {
	return fmt.Sprintf(`
data "aws_elastic_beanstalk_hosted_zone" "test" {
  region = "%s"
}
`, r)
}
