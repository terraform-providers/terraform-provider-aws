package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDataSourceElasticBeanstalkHostedZone_basic(t *testing.T) {
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsElasticBeanstalkHostedZone(dataSourceName, testAccGetRegion()),
				),
			},
		},
	})
}

func TestAccAWSDataSourceElasticBeanstalkHostedZone_Region(t *testing.T) {
	dataSourceName := "data.aws_elastic_beanstalk_hosted_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("ap-southeast-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsElasticBeanstalkHostedZone(dataSourceName, "ap-southeast-2"),
				),
			},
			{
				Config: testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("eu-west-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsElasticBeanstalkHostedZone(dataSourceName, "eu-west-1"),
				),
			},
			{
				Config:      testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion("ss-pluto-1"),
				ExpectError: regexp.MustCompile("Unsupported region"),
			},
		},
	})
}

func testAccCheckAwsElasticBeanstalkHostedZone(resourceName string, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue, ok := elasticBeanstalkHostedZoneIds[region]

		if !ok {
			return fmt.Errorf("Unsupported region: %s", region)
		}

		return resource.TestCheckResourceAttr(resourceName, "id", expectedValue)(s)
	}
}

const testAccCheckAwsElasticBeanstalkHostedZoneDataSource_currentRegion = `
data "aws_elastic_beanstalk_hosted_zone" "test" {}
`

func testAccCheckAwsElasticBeanstalkHostedZoneDataSource_byRegion(r string) string {
	return fmt.Sprintf(`
data "aws_elastic_beanstalk_hosted_zone" "test" {
  region = "%s"
}
`, r)
}
