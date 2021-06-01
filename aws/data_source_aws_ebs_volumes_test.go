package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEbsVolumes_basic(t *testing.T) {
	rInt := acctest.RandIntRange(0, 256)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEbsVolumeIDsConfig(rInt),
			},
			{
				Config: testAccDataSourceAwsEbsVolumeIDsConfigWithDataSource(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ebs_volumes.subject_under_test", "ids.#", "2"),
				),
			},
			{
				// Force the destroy to not refresh the data source (leading to an error)
				Config: testAccDataSourceAwsEbsVolumeIDsConfig(rInt),
			},
		},
	})
}

func testAccDataSourceAwsEbsVolumeIDsConfigWithDataSource(rInt int) string {
	return fmt.Sprintf(`
%s

data "aws_ebs_volumes" "subject_under_test" {
  tags = {
    TestIdentifierSet = "testAccDataSourceAwsEbsVolumes-%d"
  }
}
`, testAccDataSourceAwsEbsVolumeIDsConfig(rInt), rInt)
}

func testAccDataSourceAwsEbsVolumeIDsConfig(rInt int) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "volume" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    TestIdentifierSet = "testAccDataSourceAwsEbsVolumes-%d"
  }
}
`, rInt)
}
