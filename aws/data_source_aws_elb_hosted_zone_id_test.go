package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSElbHostedZoneId_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElbHostedZoneIdBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.main", "id", "Z1H1FL5HABSF5"),
				),
			},
			{
				Config: testAccCheckAwsElbHostedZoneIdExplicitConfig("us-east-1", "application"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "Z35SXDOTRQ7X7K"),
				),
			},
			{
				Config: testAccCheckAwsElbHostedZoneIdExplicitConfig("us-west-1", "classic"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "Z368ELLRRE2KJ0"),
				),
			},
			{
				Config: testAccCheckAwsElbHostedZoneIdExplicitConfig("eu-west-2", "network"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "ZD4D7Y8KGAS4G"),
				),
			},
		},
	})
}

const testAccCheckAwsElbHostedZoneIdBasicConfig = `
data "aws_elb_hosted_zone_id" "main" {
}
`

func testAccCheckAwsElbHostedZoneIdExplicitConfig(region, elbType string) string {
	return fmt.Sprintf(`
data "aws_elb_hosted_zone_id" "regional" {
	region = "%s"
	elb_type = "%s"
}
`, region, elbType)
}
